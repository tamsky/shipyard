package clients

import (
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"
	"github.com/shipyard-run/shipyard/pkg/utils"
	"golang.org/x/xerrors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/repo"
)

var helmLock sync.Mutex
var helmStorage = &repo.File{}

func init() {
	// create a global lock as it seems map write in Helm is not thread safe
	helmLock = sync.Mutex{}
}

// Helm defines an interface for a client which can manage Helm charts
type Helm interface {
	// CreateFromRepository creates a Helm install from a repository
	Create(kubeConfig, name, namespace string, createNamespace bool, skipCRDs bool, chart, version, valuesPath string, valuesString map[string]string) error

	// Destroy the given chart
	Destroy(kubeConfig, name, namespace string) error

	//UpsertChartRepository configures the remote chart repository
	UpsertChartRepository(name, url string) error
}

type HelmImpl struct {
	log        hclog.Logger
	repoPath   string
	cachePath  string
	dataPath   string
	configPath string
}

func NewHelm(l hclog.Logger) Helm {
	helmCachePath := path.Join(utils.GetHelmLocalFolder(""), "cache")
	helmRepoConfig := path.Join(utils.GetHelmLocalFolder(""), "repo")

	helmDataPath := path.Join(utils.GetHelmLocalFolder(""), "data")
	helmConfigPath := path.Join(utils.GetHelmLocalFolder(""), "config")

	// create the paths
	os.MkdirAll(utils.GetHelmLocalFolder(""), os.ModePerm)
	os.MkdirAll(helmCachePath, os.ModePerm)
	os.MkdirAll(helmDataPath, os.ModePerm)

	//	create the config file
	os.Create(helmRepoConfig)

	os.Setenv("HELM_CACHE_HOME", helmCachePath)
	os.Setenv("HELM_CONFIG_HOME", helmConfigPath)
	os.Setenv("HELM_DATA_HOME", helmDataPath)

	// try to load the default config
	helmStorage, _ = repo.LoadFile(helmRepoConfig)

	return &HelmImpl{l, helmRepoConfig, helmCachePath, helmDataPath, helmConfigPath}
}

func (h *HelmImpl) Create(kubeConfig, name, namespace string, createNamespace bool, skipCRDs bool, chart, version, valuesPath string, valuesString map[string]string) error {
	// set the kubeclient for Helm
	s := kube.GetConfig(kubeConfig, "default", namespace)
	cfg := &action.Configuration{}
	err := cfg.Init(s, namespace, "", func(format string, v ...interface{}) {
		h.log.Debug("Helm debug", "name", name, "chart", chart, "message", fmt.Sprintf(format, v...))
	})

	if err != nil {
		return xerrors.Errorf("unable to initialize Helm: %w", err)
	}

	client := action.NewInstall(cfg)
	client.ReleaseName = name
	client.Namespace = namespace
	client.CreateNamespace = createNamespace
	client.SkipCRDs = skipCRDs

	settings := h.getSettings()
	settings.Debug = true

	h.log.Debug("Creating chart from config", "ref", name, "chart", chart)
	cpa := client.ChartPathOptions
	cpa.Version = version

	cp, err := cpa.LocateChart(chart, &settings)
	if err != nil {
		return xerrors.Errorf("Error locating chart: %w", err)
	}

	p := getter.All(&settings)
	vo := values.Options{}
	vo.StringValues = []string{}

	// add the string values to the collection
	for k, v := range valuesString {
		vo.StringValues = append(vo.StringValues, fmt.Sprintf("%s=%s", k, v))
	}

	// if we have an overridden values file set it
	if valuesPath != "" {
		vo.ValueFiles = []string{valuesPath}
	}

	vals, err := vo.MergeValues(p)
	if err != nil {
		return xerrors.Errorf("Error merging Helm values: %w", err)
	}

	h.log.Debug("Using Values", "ref", name, "values", vals)

	h.log.Debug("Loading chart", "ref", name, "path", cp)
	chartRequested, err := loader.Load(cp)
	if err != nil {
		return xerrors.Errorf("Error loading chart: %w", err)
	}

	if err := checkIfInstallable(chartRequested); err != nil {
		return xerrors.Errorf("Chart is not installable: %w", err)
	}

	if req := chartRequested.Metadata.Dependencies; req != nil {
		h.log.Debug("Checking chart dependencies", "deps", req)

		if err := action.CheckDependencies(chartRequested, req); err != nil {
			if client.DependencyUpdate {
				man := &downloader.Manager{
					Out:              h.log.StandardWriter(&hclog.StandardLoggerOptions{}),
					ChartPath:        cp,
					Keyring:          client.ChartPathOptions.Keyring,
					SkipUpdate:       false,
					Getters:          p,
					RepositoryConfig: settings.RepositoryConfig,
					RepositoryCache:  settings.RepositoryCache,
					Debug:            h.log.IsDebug(),
				}
				if err := man.Update(); err != nil {
					return err
				}

				if chartRequested, err = loader.Load(cp); err != nil {
					return xerrors.Errorf("Failed reloading chart after repo update: %w", err)
				}
			} else {
				return err
			}
		}
	}

	h.log.Debug("Validate chart", "ref", name)
	err = chartRequested.Validate()
	if err != nil {
		return xerrors.Errorf("Error validating chart: %w", err)
	}

	h.log.Debug("Run chart", "ref", name)
	_, err = client.Run(chartRequested, vals)
	if err != nil {
		return xerrors.Errorf("Error running chart: %w", err)
	}

	return nil
}

func checkIfInstallable(ch *chart.Chart) error {
	switch ch.Metadata.Type {
	case "", "application":
		return nil
	}
	return errors.Errorf("%s charts are not installable", ch.Metadata.Type)
}

// Destroy removes an installed Helm chart from the system
func (h *HelmImpl) Destroy(kubeConfig, name, namespace string) error {
	s := kube.GetConfig(kubeConfig, "default", namespace)
	cfg := &action.Configuration{}
	err := cfg.Init(s, namespace, "", func(format string, v ...interface{}) {
		h.log.Debug("Helm debug message", "message", fmt.Sprintf(format, v...))
	})

	//settings := cli.EnvSettings{}
	//p := getter.All(&settings)
	//vo := values.Options{}
	client := action.NewUninstall(cfg)
	_, err = client.Run(name)
	if err != nil {
		h.log.Debug("Unable to remove chart, exit silently", "err", err)
		return err
	}

	return nil
}

func (h *HelmImpl) UpsertChartRepository(name, url string) error {
	r := repo.Entry{
		Name:                  name,
		URL:                   url,
		InsecureSkipTLSverify: true,
	}

	// ensure only a single client can operate at one time
	helmLock.Lock()
	defer helmLock.Unlock()

	// nothing to do
	if helmStorage.Has(r.Name) {
		return nil
	}

	settings := h.getSettings()
	p := getter.All(&settings)

	chartRepo, err := repo.NewChartRepository(&r, p)
	if err != nil {
		return fmt.Errorf("unable to create helm chart repository: %s", err)
	}

	chartRepo.CachePath = h.cachePath

	_, err = chartRepo.DownloadIndexFile()
	if err != nil {
		return fmt.Errorf("unable to download index for Helm chart: %s, %s", url, err)
	}

	helmStorage.Update(&r)
	err = helmStorage.WriteFile(settings.RepositoryConfig, 0644)
	if err != nil {
		return fmt.Errorf("unable to update Helm storage: %s", err)
	}

	return nil
}

func (h *HelmImpl) getSettings() cli.EnvSettings {
	settings := cli.EnvSettings{}
	settings.RepositoryConfig = h.repoPath
	settings.RepositoryCache = h.cachePath

	return settings
}
