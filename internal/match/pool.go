package match

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"lightning-engine/conf"
	"lightning-engine/internal/status"
	mainlog "lightning-engine/log"
	"lightning-engine/models"
	"lightning-engine/models/dex_model"
	"lightning-engine/mq"
)

// MatchPool 撮合池
type MatchPool struct {
	pool map[string]*Orderbook
	DB   *gorm.DB
}

func connectToPostgreSQL() (*gorm.DB, error) {
	//dsn := "user=hi-five password=123456 dbname=hi-five host=192.168.80.87 port=5432 sslmode=disable"
	//dsn := "user=postgres password=123456 dbname=hi-five host=localhost port=5432 sslmode=disable"
	dsn := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		conf.Gconfig.GetString("DB.user"),
		conf.Gconfig.GetString("DB.password"),
		conf.Gconfig.GetString("DB.dbname"),
		conf.Gconfig.GetString("DB.host"),
		conf.Gconfig.GetString("DB.port"))
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func NewMatchPool(status *status.Status, pairs []string, mq mq.IMQ) (*MatchPool, error) {
	db, err := connectToPostgreSQL()
	if err != nil {
		mainlog.Info("[NewMatchPool] init db err:%+v\n", err)
		panic(err)
	}

	mp := MatchPool{}
	mp.pool = make(map[string]*Orderbook)
	mp.DB = db
	for _, p := range pairs {
		ob, err := NewOrderbook(status, p, mq)
		if err != nil {
			return nil, err
		}
		status.Add(1)
		go ob.Begin()

		// 通过数据库初始化撮合引擎
		var dexOrders []dex_model.DexOrder
		result := db.Where("symbol = ? and (status = ? or status= ?)", p, dex_model.NotTraded, dex_model.PartTraded).Find(&dexOrders)
		if result.Error != nil {
			mainlog.Info("[NewMatchPool] init db use %s err:%+v\n", p, result.Error)
		}
		for _, dexOrder := range dexOrders {
			newOrder := &models.Order{
				Id:          dexOrder.OrderID,
				UserId:      int64(dexOrder.UId),
				Pair:        p,
				Price:       dexOrder.Price,
				Amount:      dexOrder.TotalVolume.Sub(dexOrder.TradedVolume),
				Side:        models.GetOrderSideFromDexOrder(dexOrder.SideType),
				Type:        models.GetOrderTypeFromDexOrder(dexOrder.OrderType),
				TimeInForce: models.TimeInForceGTC,
			}
			mainlog.Info("now add %s -> %s\n", newOrder.Id, p)
			err := ob.Add(newOrder)
			if err != nil {
				mainlog.Info("[NewMatchPool] init db use dex_order %s err:%+v\n", dexOrder.OrderID, result.Error)
			}
		}
		mp.pool[p] = ob
	}
	return &mp, nil
}

// AddOrder 挂单
func (mp *MatchPool) AddOrder(order *models.Order) error {
	if _, ok := mp.pool[order.Pair]; !ok {
		return ErrPair
	}
	return mp.pool[order.Pair].Add(order)
}

// CancelOrder 撤单
func (mp *MatchPool) CancelOrder(pair string, id string) error {
	if _, ok := mp.pool[pair]; !ok {
		return ErrPair
	}
	return mp.pool[pair].Cancel(id)
}
