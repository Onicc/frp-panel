package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/Onicc/frp-panel/models"
	"github.com/Onicc/frp-panel/pb"
	"github.com/Onicc/frp-panel/utils"
	"github.com/Onicc/frp-panel/utils/logger"
	"github.com/samber/lo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProxyQuery interface {
	GetProxyStatsByClientID(userInfo models.UserInfo, clientID string) ([]*models.ProxyStatsEntity, error)
	GetProxyStatsByServerID(userInfo models.UserInfo, serverID string) ([]*models.ProxyStatsEntity, error)
	AdminGetTenantProxyStats(tenantID int) ([]*models.ProxyStatsEntity, error)
	AdminGetAllProxyStats(tx *gorm.DB) ([]*models.ProxyStatsEntity, error)
	AdminGetProxyConfigByClientIDAndName(clientID string, name string) (*models.ProxyConfig, error)
	GetProxyConfigsByClientID(userInfo models.UserInfo, clientID string) ([]*models.ProxyConfigEntity, error)
	GetProxyConfigByFilter(userInfo models.UserInfo, proxyConfig *models.ProxyConfigEntity) (*models.ProxyConfig, error)
	ListProxyConfigsWithFilters(userInfo models.UserInfo, page, pageSize int, filters *models.ProxyConfigEntity) ([]*models.ProxyConfig, error)
	AdminListProxyConfigsWithFilters(filters *models.ProxyConfigEntity) ([]*models.ProxyConfig, error)
	ListProxyConfigsWithFiltersAndKeyword(userInfo models.UserInfo, page, pageSize int, filters *models.ProxyConfigEntity, keyword string) ([]*models.ProxyConfig, error)
	ListProxyConfigsWithKeyword(userInfo models.UserInfo, page, pageSize int, keyword string) ([]*models.ProxyConfig, error)
	ListProxyConfigs(userInfo models.UserInfo, page, pageSize int) ([]*models.ProxyConfig, error)
	GetProxyConfigByOriginClientIDAndName(userInfo models.UserInfo, clientID string, name string) (*models.ProxyConfig, error)
	CountProxyConfigs(userInfo models.UserInfo) (int64, error)
	CountProxyConfigsWithFilters(userInfo models.UserInfo, filters *models.ProxyConfigEntity) (int64, error)
	CountProxyConfigsWithFiltersAndKeyword(userInfo models.UserInfo, filters *models.ProxyConfigEntity, keyword string) (int64, error)
	GetProxyConfigsByWorkerId(userInfo models.UserInfo, workerID string) ([]*models.ProxyConfig, error)
}

type ProxyMutation interface {
	AdminUpdateProxyStats(srv *models.ServerEntity, inputs []*pb.ProxyInfo) error
	AdminCreateProxyConfig(proxyCfg *models.ProxyConfig) error
	RebuildProxyConfigFromClient(userInfo models.UserInfo, client *models.Client) error
	CreateProxyConfig(userInfo models.UserInfo, proxyCfg *models.ProxyConfigEntity) error
	UpdateProxyConfig(userInfo models.UserInfo, proxyCfg *models.ProxyConfig) error
	DeleteProxyConfig(userInfo models.UserInfo, clientID, name string) error
	DeleteProxyConfigsByClientIDOrOriginClientID(userInfo models.UserInfo, clientID string) error
	DeleteProxyConfigsByClientID(userInfo models.UserInfo, clientID string) error
}

type proxyQuery struct{ *queryImpl }
type proxyMutation struct{ *mutationImpl }

func newProxyQuery(base *queryImpl) ProxyQuery          { return &proxyQuery{base} }
func newProxyMutation(base *mutationImpl) ProxyMutation { return &proxyMutation{base} }

func (q *proxyQuery) GetProxyStatsByClientID(userInfo models.UserInfo, clientID string) ([]*models.ProxyStatsEntity, error) {
	if clientID == "" {
		return nil, fmt.Errorf("invalid client id")
	}
	db := q.ctx.GetApp().GetDBManager().GetDefaultDB()
	list := []*models.ProxyStats{}
	err := db.
		Where(&models.ProxyStats{ProxyStatsEntity: &models.ProxyStatsEntity{
			UserID:   userInfo.GetUserID(),
			TenantID: userInfo.GetTenantID(),
			ClientID: clientID,
		}}).
		Or(&models.ProxyStats{ProxyStatsEntity: &models.ProxyStatsEntity{
			UserID:   0,
			TenantID: userInfo.GetTenantID(),
			ClientID: clientID,
		}}).
		Or(&models.ProxyStats{ProxyStatsEntity: &models.ProxyStatsEntity{
			UserID:         userInfo.GetUserID(),
			TenantID:       userInfo.GetTenantID(),
			OriginClientID: clientID,
		}}).
		Or(&models.ProxyStats{ProxyStatsEntity: &models.ProxyStatsEntity{
			UserID:         0,
			TenantID:       userInfo.GetTenantID(),
			OriginClientID: clientID,
		}}).
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return lo.Map(list, func(item *models.ProxyStats, _ int) *models.ProxyStatsEntity {
		return item.ProxyStatsEntity
	}), nil
}

func (q *proxyQuery) GetProxyStatsByServerID(userInfo models.UserInfo, serverID string) ([]*models.ProxyStatsEntity, error) {
	if serverID == "" {
		return nil, fmt.Errorf("invalid server id")
	}
	db := q.ctx.GetApp().GetDBManager().GetDefaultDB()
	list := []*models.ProxyStats{}
	err := db.
		Where(&models.ProxyStats{ProxyStatsEntity: &models.ProxyStatsEntity{
			UserID:   userInfo.GetUserID(),
			TenantID: userInfo.GetTenantID(),
			ServerID: serverID,
		}}).Or(&models.ProxyStats{ProxyStatsEntity: &models.ProxyStatsEntity{
		UserID:   0,
		TenantID: userInfo.GetTenantID(),
		ServerID: serverID,
	}}).
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return lo.Map(list, func(item *models.ProxyStats, _ int) *models.ProxyStatsEntity {
		return item.ProxyStatsEntity
	}), nil
}

func (m *proxyMutation) AdminUpdateProxyStats(srv *models.ServerEntity, inputs []*pb.ProxyInfo) error {
	if srv.ServerID == "" {
		return fmt.Errorf("invalid server id")
	}

	db := m.ctx.GetApp().GetDBManager().GetDefaultDB()
	return db.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Where("user_id = ?", srv.UserID).First(&user).Error; err != nil {
			return err
		}
		var tunnels []models.ProxyConfig
		if err := tx.Where("user_id = ? AND server_id = ? AND managed_by = ?", srv.UserID, srv.ServerID, "tunnel").Find(&tunnels).Error; err != nil {
			return err
		}
		var clients []models.Client
		if err := tx.Where("user_id = ? AND server_id = ?", srv.UserID, srv.ServerID).Find(&clients).Error; err != nil {
			return err
		}
		var previous []models.ProxyStats
		if err := tx.Where("user_id = ? AND server_id = ?", srv.UserID, srv.ServerID).Find(&previous).Error; err != nil {
			return err
		}
		type identity struct{ clientID, originID, name string }
		byWireName := make(map[string]identity, len(tunnels))
		for _, tunnel := range tunnels {
			if tunnel.Stopped {
				continue
			}
			byWireName[utils.FRPClientUser(user.UserName, tunnel.OriginClientID)+"."+tunnel.Name] = identity{tunnel.OriginClientID, tunnel.OriginClientID, tunnel.Name}
		}
		for _, client := range clients {
			cfg, err := client.GetConfigContent()
			if err != nil || cfg == nil {
				continue
			}
			for _, proxy := range cfg.Proxies {
				name := proxy.GetBaseConfig().Name
				byWireName[cfg.User+"."+name] = identity{client.ClientID, client.OriginClientID, name}
			}
		}
		oldByIdentity := make(map[string]*models.ProxyStats, len(previous))
		for i := range previous {
			old := &previous[i]
			oldByIdentity[old.ClientID+"\x00"+old.Name] = old
		}
		results := make([]*models.ProxyStats, 0, len(inputs))
		now := time.Now()
		for _, proxyInfo := range inputs {
			if proxyInfo == nil {
				continue
			}
			key, ok := byWireName[proxyInfo.GetName()]
			if !ok {
				continue
			}
			item := &models.ProxyStats{ProxyStatsEntity: &models.ProxyStatsEntity{
				ServerID:        srv.ServerID,
				ClientID:        key.clientID,
				OriginClientID:  key.originID,
				Name:            key.name,
				Type:            proxyInfo.GetType(),
				UserID:          srv.UserID,
				TenantID:        srv.TenantID,
				TodayTrafficIn:  proxyInfo.GetTodayTrafficIn(),
				TodayTrafficOut: proxyInfo.GetTodayTrafficOut(),
			}}
			if old := oldByIdentity[key.clientID+"\x00"+key.name]; old != nil {
				item.ProxyID = old.ProxyID
				item.HistoryTrafficIn = old.HistoryTrafficIn
				item.HistoryTrafficOut = old.HistoryTrafficOut
				if !utils.IsSameDay(now, old.UpdatedAt) || proxyInfo.GetFirstSync() {
					item.HistoryTrafficIn += old.TodayTrafficIn
					item.HistoryTrafficOut += old.TodayTrafficOut
				}
			}
			results = append(results, item)
		}

		if len(results) > 0 {
			return tx.Save(results).Error
		}
		return nil
	})
}

func (q *proxyQuery) AdminGetTenantProxyStats(tenantID int) ([]*models.ProxyStatsEntity, error) {
	db := q.ctx.GetApp().GetDBManager().GetDefaultDB()
	list := []*models.ProxyStats{}
	err := db.
		Where(&models.ProxyStats{ProxyStatsEntity: &models.ProxyStatsEntity{
			TenantID: tenantID,
		}}).
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return lo.Map(list, func(item *models.ProxyStats, _ int) *models.ProxyStatsEntity {
		return item.ProxyStatsEntity
	}), nil
}

func (q *proxyQuery) AdminGetAllProxyStats(tx *gorm.DB) ([]*models.ProxyStatsEntity, error) {
	db := tx
	list := []*models.ProxyStats{}
	err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return lo.Map(list, func(item *models.ProxyStats, _ int) *models.ProxyStatsEntity {
		return item.ProxyStatsEntity
	}), nil
}

func (m *proxyMutation) AdminCreateProxyConfig(proxyCfg *models.ProxyConfig) error {
	db := m.ctx.GetApp().GetDBManager().GetDefaultDB()
	return db.Create(proxyCfg).Error
}

// RebuildProxyConfigFromClient rebuild proxy from client
// skip stopped proxy
func (m *proxyMutation) RebuildProxyConfigFromClient(userInfo models.UserInfo, client *models.Client) error {
	db := m.ctx.GetApp().GetDBManager().GetDefaultDB()
	query := NewQuery(m.ctx)

	pxyCfgs, err := utils.LoadProxiesFromContent(client.ConfigContent)
	if err != nil {
		return err
	}

	proxyConfigEntities := []*models.ProxyConfig{}

	for _, pxyCfg := range pxyCfgs {
		proxyCfg := &models.ProxyConfig{
			ProxyConfigEntity: &models.ProxyConfigEntity{},
		}
		if oldProxyCfg, err := query.GetProxyConfigByOriginClientIDAndName(userInfo, client.ClientID, pxyCfg.GetBaseConfig().Name); err == nil {
			logger.Logger(context.Background()).WithError(err).Warnf("proxy config already exist, will be override, clientID: [%s], name: [%s]",
				client.ClientID, pxyCfg.GetBaseConfig().Name)
			proxyCfg.Model = oldProxyCfg.Model
		}

		if err := proxyCfg.FillClientConfig(client.ClientEntity); err != nil {
			return err
		}

		if err := proxyCfg.FillTypedProxyConfig(pxyCfg); err != nil {
			return err
		}

		proxyConfigEntities = append(proxyConfigEntities, proxyCfg)
	}

	if err := m.DeleteProxyConfigsByClientIDOrOriginClientID(userInfo, client.ClientID); err != nil {
		return err
	}

	if len(proxyConfigEntities) == 0 {
		return nil
	}

	return db.Save(proxyConfigEntities).Error
}

func (q *proxyQuery) AdminGetProxyConfigByClientIDAndName(clientID string, name string) (*models.ProxyConfig, error) {
	db := q.ctx.GetApp().GetDBManager().GetDefaultDB()
	proxyCfg := &models.ProxyConfig{}
	err := db.
		Where(&models.ProxyConfig{ProxyConfigEntity: &models.ProxyConfigEntity{
			ClientID: clientID,
			Name:     name,
		}}).
		First(proxyCfg).Error
	if err != nil {
		return nil, err
	}
	return proxyCfg, nil
}

func (q *proxyQuery) GetProxyConfigsByClientID(userInfo models.UserInfo, clientID string) ([]*models.ProxyConfigEntity, error) {
	if clientID == "" {
		return nil, fmt.Errorf("invalid client id")
	}
	db := q.ctx.GetApp().GetDBManager().GetDefaultDB()
	list := []*models.ProxyConfig{}
	err := db.
		Where(&models.ProxyConfig{ProxyConfigEntity: &models.ProxyConfigEntity{
			UserID:   userInfo.GetUserID(),
			TenantID: userInfo.GetTenantID(),
			ClientID: clientID,
		}}).
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return lo.Map(list, func(item *models.ProxyConfig, _ int) *models.ProxyConfigEntity {
		return item.ProxyConfigEntity
	}), nil
}

func (q *proxyQuery) GetProxyConfigByFilter(userInfo models.UserInfo, proxyConfig *models.ProxyConfigEntity) (*models.ProxyConfig, error) {
	db := q.ctx.GetApp().GetDBManager().GetDefaultDB()
	filter := &models.ProxyConfigEntity{}

	if len(proxyConfig.ClientID) != 0 {
		filter.ClientID = proxyConfig.ClientID
	}
	if len(proxyConfig.OriginClientID) != 0 {
		filter.OriginClientID = proxyConfig.OriginClientID
	}
	if len(proxyConfig.Name) != 0 {
		filter.Name = proxyConfig.Name
	}
	if len(proxyConfig.Type) != 0 {
		filter.Type = proxyConfig.Type
	}
	if len(proxyConfig.ServerID) != 0 {
		filter.ServerID = proxyConfig.ServerID
	}

	filter.UserID = userInfo.GetUserID()
	filter.TenantID = userInfo.GetTenantID()

	respProxyCfg := &models.ProxyConfig{}
	err := db.
		Where(&models.ProxyConfig{ProxyConfigEntity: filter}).
		First(respProxyCfg).Error
	if err != nil {
		return nil, err
	}
	return respProxyCfg, nil
}

func (q *proxyQuery) ListProxyConfigsWithFilters(userInfo models.UserInfo, page, pageSize int, filters *models.ProxyConfigEntity) ([]*models.ProxyConfig, error) {
	if page < 1 || pageSize < 1 {
		return nil, fmt.Errorf("invalid page or page size")
	}

	db := q.ctx.GetApp().GetDBManager().GetDefaultDB()
	offset := (page - 1) * pageSize

	filters.UserID = userInfo.GetUserID()
	filters.TenantID = userInfo.GetTenantID()

	var proxyConfigs []*models.ProxyConfig
	err := db.Where(&models.ProxyConfig{
		ProxyConfigEntity: filters,
	}).Where(filters).Offset(offset).Limit(pageSize).Find(&proxyConfigs).Error
	if err != nil {
		return nil, err
	}

	return proxyConfigs, nil
}

func (q *proxyQuery) AdminListProxyConfigsWithFilters(filters *models.ProxyConfigEntity) ([]*models.ProxyConfig, error) {
	db := q.ctx.GetApp().GetDBManager().GetDefaultDB()

	var proxyConfigs []*models.ProxyConfig
	err := db.Where(&models.ProxyConfig{
		ProxyConfigEntity: filters,
	}).Where(filters).Find(&proxyConfigs).Error
	if err != nil {
		return nil, err
	}

	return proxyConfigs, nil
}

func (q *proxyQuery) ListProxyConfigsWithFiltersAndKeyword(userInfo models.UserInfo, page, pageSize int, filters *models.ProxyConfigEntity, keyword string) ([]*models.ProxyConfig, error) {
	if page < 1 || pageSize < 1 || len(keyword) == 0 {
		return nil, fmt.Errorf("invalid page or page size or keyword")
	}

	db := q.ctx.GetApp().GetDBManager().GetDefaultDB()
	offset := (page - 1) * pageSize

	filters.UserID = userInfo.GetUserID()
	filters.TenantID = userInfo.GetTenantID()

	var proxyConfigs []*models.ProxyConfig
	err := db.Where(&models.ProxyConfig{
		ProxyConfigEntity: filters,
	}).Where(filters).Where("name like ?", "%"+keyword+"%").Offset(offset).Limit(pageSize).Find(&proxyConfigs).Error
	if err != nil {
		return nil, err
	}

	return proxyConfigs, nil
}

func (q *proxyQuery) ListProxyConfigsWithKeyword(userInfo models.UserInfo, page, pageSize int, keyword string) ([]*models.ProxyConfig, error) {
	return q.ListProxyConfigsWithFiltersAndKeyword(userInfo, page, pageSize, &models.ProxyConfigEntity{}, keyword)
}

func (q *proxyQuery) ListProxyConfigs(userInfo models.UserInfo, page, pageSize int) ([]*models.ProxyConfig, error) {
	return q.ListProxyConfigsWithFilters(userInfo, page, pageSize, &models.ProxyConfigEntity{})
}

func (m *proxyMutation) CreateProxyConfig(userInfo models.UserInfo, proxyCfg *models.ProxyConfigEntity) error {
	db := m.ctx.GetApp().GetDBManager().GetDefaultDB()
	proxyCfg.UserID = userInfo.GetUserID()
	proxyCfg.TenantID = userInfo.GetTenantID()
	return db.Create(&models.ProxyConfig{ProxyConfigEntity: proxyCfg}).Error
}

func (m *proxyMutation) UpdateProxyConfig(userInfo models.UserInfo, proxyCfg *models.ProxyConfig) error {
	if proxyCfg.ID == 0 {
		return fmt.Errorf("invalid proxy config id")
	}
	db := m.ctx.GetApp().GetDBManager().GetDefaultDB()
	proxyCfg.UserID = userInfo.GetUserID()
	proxyCfg.TenantID = userInfo.GetTenantID()
	return db.Where(&models.ProxyConfig{
		ProxyConfigEntity: &models.ProxyConfigEntity{
			UserID:   userInfo.GetUserID(),
			TenantID: userInfo.GetTenantID(),
			ClientID: proxyCfg.ClientID,
		},
		Model: &gorm.Model{
			ID: proxyCfg.ID,
		},
	}).Save(proxyCfg).Error
}

func (m *proxyMutation) DeleteProxyConfig(userInfo models.UserInfo, clientID, name string) error {
	if clientID == "" || name == "" {
		return fmt.Errorf("invalid client id or name")
	}
	db := m.ctx.GetApp().GetDBManager().GetDefaultDB()
	return db.Unscoped().
		Where(&models.ProxyConfig{ProxyConfigEntity: &models.ProxyConfigEntity{
			UserID:   userInfo.GetUserID(),
			TenantID: userInfo.GetTenantID(),
			ClientID: clientID,
			Name:     name,
		}}).
		Delete(&models.ProxyConfig{}).Error
}

func (m *proxyMutation) DeleteProxyConfigsByClientIDOrOriginClientID(userInfo models.UserInfo, clientID string) error {
	if clientID == "" {
		return fmt.Errorf("invalid client id")
	}
	db := m.ctx.GetApp().GetDBManager().GetDefaultDB()
	return db.Unscoped().
		Where(
			db.Where(&models.ProxyConfig{ProxyConfigEntity: &models.ProxyConfigEntity{
				UserID:   userInfo.GetUserID(),
				TenantID: userInfo.GetTenantID(),
				ClientID: clientID,
			}}).
				Or(&models.ProxyConfig{ProxyConfigEntity: &models.ProxyConfigEntity{
					UserID:         userInfo.GetUserID(),
					TenantID:       userInfo.GetTenantID(),
					OriginClientID: clientID,
				}})).
		Where(db.Where("stopped is NULL").
			Or("stopped = ?", false)).
		Delete(&models.ProxyConfig{}).Error
}

func (m *proxyMutation) DeleteProxyConfigsByClientID(userInfo models.UserInfo, clientID string) error {
	if clientID == "" {
		return fmt.Errorf("invalid client id")
	}
	db := m.ctx.GetApp().GetDBManager().GetDefaultDB()
	return db.Unscoped().
		Where(&models.ProxyConfig{ProxyConfigEntity: &models.ProxyConfigEntity{
			UserID:   userInfo.GetUserID(),
			TenantID: userInfo.GetTenantID(),
			ClientID: clientID,
		}}).
		Delete(&models.ProxyConfig{}).Error
}

func (q *proxyQuery) GetProxyConfigByOriginClientIDAndName(userInfo models.UserInfo, clientID string, name string) (*models.ProxyConfig, error) {
	if clientID == "" || name == "" {
		return nil, fmt.Errorf("invalid client id or name")
	}
	db := q.ctx.GetApp().GetDBManager().GetDefaultDB()
	item := &models.ProxyConfig{}
	err := db.
		Where(&models.ProxyConfig{ProxyConfigEntity: &models.ProxyConfigEntity{
			UserID:         userInfo.GetUserID(),
			TenantID:       userInfo.GetTenantID(),
			OriginClientID: clientID,
			Name:           name,
		}}).
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (q *proxyQuery) CountProxyConfigs(userInfo models.UserInfo) (int64, error) {
	return q.CountProxyConfigsWithFilters(userInfo, &models.ProxyConfigEntity{})
}

func (q *proxyQuery) CountProxyConfigsWithFilters(userInfo models.UserInfo, filters *models.ProxyConfigEntity) (int64, error) {
	db := q.ctx.GetApp().GetDBManager().GetDefaultDB()
	filters.UserID = userInfo.GetUserID()
	filters.TenantID = userInfo.GetTenantID()

	var count int64
	err := db.Model(&models.ProxyConfig{}).Where(&models.ProxyConfig{
		ProxyConfigEntity: filters,
	}).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (q *proxyQuery) CountProxyConfigsWithFiltersAndKeyword(userInfo models.UserInfo, filters *models.ProxyConfigEntity, keyword string) (int64, error) {
	if len(keyword) == 0 {
		return q.CountProxyConfigsWithFilters(userInfo, filters)
	}

	db := q.ctx.GetApp().GetDBManager().GetDefaultDB()
	filters.UserID = userInfo.GetUserID()
	filters.TenantID = userInfo.GetTenantID()

	var count int64
	err := db.Model(&models.ProxyConfig{}).Where(&models.ProxyConfig{
		ProxyConfigEntity: filters,
	}).Where("name like ?", "%"+keyword+"%").Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (q *proxyQuery) GetProxyConfigsByWorkerId(userInfo models.UserInfo, workerID string) ([]*models.ProxyConfig, error) {
	db := q.ctx.GetApp().GetDBManager().GetDefaultDB()
	items := []*models.ProxyConfig{}

	err := db.
		Where(&models.ProxyConfig{ProxyConfigEntity: &models.ProxyConfigEntity{
			UserID:   userInfo.GetUserID(),
			TenantID: userInfo.GetTenantID(),
		},
			WorkerID: workerID,
		}).
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}
