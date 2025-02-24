package service

import (
	"context"
	"github.com/esnet/gdg/internal/api"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/grafana-swagger-api-golang/goclient/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"sync"
)

type GrafanaService interface {
	OrganizationsApi
	DashboardsApi
	ConnectionsApi
	AlertNotificationsApi
	UsersApi
	FoldersApi
	LibraryElementsApi
	TeamsApi

	AuthenticationApi
	//MetaData
	GetServerInfo() map[string]interface{}
}

var (
	instance        *DashNGoImpl
	initServiceOnce sync.Once
)

type DashNGoImpl struct {
	client   *client.GrafanaHTTPAPI
	extended *api.ExtendedApi

	grafanaConf *config.GrafanaConfig
	configRef   *viper.Viper
	debug       bool
	storage     Storage
}

func NewDashNGoImpl() *DashNGoImpl {
	initServiceOnce.Do(func() {
		instance = newInstance()
	})
	return instance
}

func newInstance() *DashNGoImpl {
	obj := &DashNGoImpl{}
	obj.grafanaConf = config.Config().GetDefaultGrafanaConfig()
	obj.configRef = config.Config().ViperConfig()
	obj.Login()

	obj.debug = config.Config().IsDebug()
	configureStorage(obj)

	return obj
}

// Testing Only
func (s *DashNGoImpl) SetStorage(v Storage) {
	s.storage = v
}

func configureStorage(obj *DashNGoImpl) {
	//config
	storageType, appData := config.Config().GetCloudConfiguration(config.Config().GetDefaultGrafanaConfig().Storage)

	var err error
	ctx := context.Background()
	ctx = context.WithValue(ctx, StorageContext, appData)
	switch storageType {
	case "cloud":
		{
			obj.storage, err = NewCloudStorage(ctx)
			if err != nil {
				log.Warn("falling back on Local Storage, Cloud storage configuration error")
				obj.storage = NewLocalStorage(ctx)
			}
		}
	default:
		obj.storage = NewLocalStorage(ctx)
	}
}

func NewApiService(override ...string) GrafanaService {
	//Used for Testing purposes
	if len(override) > 0 {
		return newInstance()
	}
	return NewDashNGoImpl()
}
