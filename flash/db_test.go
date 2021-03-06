package flash

import (
	ctx "context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/TerrexTech/go-commonutils/commonutil"
	mongo "github.com/TerrexTech/go-mongoutils/mongo"
	"github.com/TerrexTech/uuuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

func TestBooks(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Flash sale Suite")
}

// newTimeoutContext creates a new WithTimeout context with specified timeout.
func newTimeoutContext(timeout uint32) (ctx.Context, ctx.CancelFunc) {
	return ctx.WithTimeout(
		ctx.Background(),
		time.Duration(timeout)*time.Millisecond,
	)
}

type Env struct {
	Flashdb     DBI
	Metricdb    DBI
	Inventorydb DBI
}

type mockDb struct {
	collection *mongo.Collection
}

var _ = Describe("Mongo service test", func() {
	var (
		// jsonString string
		// mgTable         *mongo.Collection
		// client          mongo.Client
		resourceTimeout uint32
		testDatabase    string
		// clientConfig    mongo.ClientConfig
		// c               *mongo.Collection
		// dataCount int
		// unixTime        int64
		configFlash DBIConfig
		// configMetric DBIConfig
		// configInv    DBIConfig
	)

	testDatabase = "rns_flash_test"
	resourceTimeout = 3000
	// dataCount = 5

	dropTestDatabase := func(dbConfig DBIConfig) {
		clientConfig := mongo.ClientConfig{
			Hosts:               dbConfig.Hosts,
			Username:            dbConfig.Username,
			Password:            dbConfig.Password,
			TimeoutMilliseconds: dbConfig.TimeoutMilliseconds,
		}
		client, err := mongo.NewClient(clientConfig)
		Expect(err).ToNot(HaveOccurred())

		dbCtx, dbCancel := newTimeoutContext(resourceTimeout)
		err = client.Database(testDatabase).Drop(dbCtx)
		dbCancel()
		Expect(err).ToNot(HaveOccurred())

		err = client.Disconnect()
		Expect(err).ToNot(HaveOccurred())
	}

	GenerateTestDB := func(dbConfig DBIConfig, schema interface{}) (*DB, error) {
		config := mongo.ClientConfig{
			Hosts:               dbConfig.Hosts,
			Username:            dbConfig.Username,
			Password:            dbConfig.Password,
			TimeoutMilliseconds: dbConfig.TimeoutMilliseconds,
		}

		client, err := mongo.NewClient(config)
		if err != nil {
			err = errors.Wrap(err, "Error creating DB-client")
			return nil, err
		}

		conn := &mongo.ConnectionConfig{
			Client:  client,
			Timeout: 5000,
		}

		// ====> Create New Collection
		collConfig := &mongo.Collection{
			Connection:   conn,
			Database:     dbConfig.Database,
			Name:         dbConfig.Collection,
			SchemaStruct: schema,
			// Indexes:      indexConfigs,
		}
		c, err := mongo.EnsureCollection(collConfig)
		if err != nil {
			err = errors.Wrap(err, "Error creating DB-client")
			return nil, err
		}
		return &DB{
			collection: c,
		}, nil
	}

	CreateNewUUID := func() (uuuid.UUID, error) {
		uuid, err := uuuid.NewV4()
		Expect(err).ToNot(HaveOccurred())

		return uuid, nil
	}

	BeforeEach(func() {
		configFlash = DBIConfig{
			Hosts:               *commonutil.ParseHosts("localhost:27017"),
			Username:            "root",
			Password:            "root",
			TimeoutMilliseconds: 3000,
			Database:            testDatabase,
			Collection:          "agg_flash",
		}

		// configMetric = DBIConfig{
		// 	Hosts:               *commonutil.ParseHosts("localhost:27017"),
		// 	Username:            "root",
		// 	Password:            "root",
		// 	TimeoutMilliseconds: 3000,
		// 	Database:            testDatabase,
		// 	Collection:          "agg_metric",
		// }

		// configInv = DBIConfig{
		// 	Hosts:               *commonutil.ParseHosts("localhost:27017"),
		// 	Username:            "root",
		// 	Password:            "root",
		// 	TimeoutMilliseconds: 3000,
		// 	Database:            testDatabase,
		// 	Collection:          "agg_inventory",
		// }
	})

	AfterEach(func() {
		configFlash := DBIConfig{
			Hosts:               *commonutil.ParseHosts("localhost:27017"),
			Username:            "root",
			Password:            "root",
			TimeoutMilliseconds: 3000,
			Database:            "rns_projections",
			Collection:          "agg_flash",
		}

		// configMetric := DBIConfig{
		// 	Hosts:               *commonutil.ParseHosts("localhost:27017"),
		// 	Username:            "root",
		// 	Password:            "root",
		// 	TimeoutMilliseconds: 3000,
		// 	Database:            "rns_projections",
		// 	Collection:          "agg_metric",
		// }

		// configInv := DBIConfig{
		// 	Hosts:               *commonutil.ParseHosts("localhost:27017"),
		// 	Username:            "root",
		// 	Password:            "root",
		// 	TimeoutMilliseconds: 3000,
		// 	Database:            "rns_projections",
		// 	Collection:          "agg_inventory",
		// }
		dropTestDatabase(configFlash)
		// err := client.Disconnect()
		// Expect(err).ToNot(HaveOccurred())

		// dropTestDatabase(configMetric)
		// err = client.Disconnect()
		// Expect(err).ToNot(HaveOccurred())

		// dropTestDatabase(configInv)
		// err = client.Disconnect()
		// Expect(err).ToNot(HaveOccurred())
	})

	It("Delete flashsale", func() {

		dbFlash, err := GenerateTestDB(configFlash, &Flash{})
		Expect(err).ToNot(HaveOccurred())

		// _, err = GenerateTestDB(configMetric, &Metric{})
		// Expect(err).ToNot(HaveOccurred())

		// _, err = GenerateTestDB(configInv, &Inventory{})
		// Expect(err).ToNot(HaveOccurred())

		flashId, err := CreateNewUUID()
		Expect(err).ToNot(HaveOccurred())

		itemId, err := CreateNewUUID()
		Expect(err).ToNot(HaveOccurred())

		deviceId, err := CreateNewUUID()
		Expect(err).ToNot(HaveOccurred())

		fSaleData := fmt.Sprintf(`[{"flash_id": "%v","item_id":"%v","upc":222222222,"sku":343434,"name":"test","origin":"ON Canada","device_id":"%v","price":100,"sale_price": 10,"ethylene":800,"status":"active"}]`, flashId, itemId, deviceId)

		flashSale := []Flash{}

		err = json.Unmarshal([]byte(fSaleData), &flashSale)
		Expect(err).ToNot(HaveOccurred())

		uuid, err := uuuid.NewV4()
		Expect(err).ToNot(HaveOccurred())

		for _, v := range flashSale {
			v.FlashID = uuid
			v.Timestamp = time.Now().Unix()

			_, err = dbFlash.collection.InsertOne(&v)
			Expect(err).ToNot(HaveOccurred())
		}

		fsaleDelete := fmt.Sprintf(`[{"flash_id": %d}]`, flashId)

		flashDelete := []Flash{}

		err = json.Unmarshal([]byte(fsaleDelete), &flashDelete)
		Expect(err).ToNot(HaveOccurred())

		deleteResult, err := dbFlash.DeleteFlashSale(flashDelete)
		Expect(err).ToNot(HaveOccurred())
		for _, v := range deleteResult {
			Expect(v.DeletedCount).To(Equal(int64(1)))
		}

		for _, v := range flashDelete {
			_, err = dbFlash.collection.Find(map[string]interface{}{
				"flash_id": map[string]interface{}{
					"eq": &v.FlashID,
				},
			})
			Expect(err).To(HaveOccurred())
		}
	})

	// It("Should give an error if item_id is missing", func() {
	// 	dbFlash, err := GenerateTestDB(configFlash, &Flash{})
	// 	Expect(err).ToNot(HaveOccurred())

	// 	flashId, err := CreateNewUUID()
	// 	Expect(err).ToNot(HaveOccurred())

	// 	deviceId, err := CreateNewUUID()
	// 	Expect(err).ToNot(HaveOccurred())

	// 	fSaleData := fmt.Sprintf(`[{"flash_id": "%v","upc":222222222,"sku":343434,"name":"test","origin":"ON Canada","device_id":"%v","price":100,"sale_price": 10,"ethylene":800}]`, flashId, deviceId)

	// 	flashSale := []Flash{}

	// 	err = json.Unmarshal([]byte(fSaleData), &flashSale)
	// 	Expect(err).ToNot(HaveOccurred())

	// 	_, err = dbFlash.UpdateFlashSale(flashSale)
	// 	Expect(err).To(HaveOccurred())
	// })

	// It("Should give an error if upc is missing", func() {
	// 	dbFlash, err := GenerateTestDB(configFlash, &Flash{})
	// 	Expect(err).ToNot(HaveOccurred())

	// 	flashId, err := CreateNewUUID()
	// 	Expect(err).ToNot(HaveOccurred())

	// 	itemId, err := CreateNewUUID()
	// 	Expect(err).ToNot(HaveOccurred())

	// 	deviceId, err := CreateNewUUID()
	// 	Expect(err).ToNot(HaveOccurred())

	// 	fSaleData := fmt.Sprintf(`[{"flash_id": "%v","item_id":"%v","sku":343434,"name":"test","origin":"ON Canada","device_id":"%v","price":100,"sale_price": 10,"ethylene":800}]`, flashId, itemId, deviceId)

	// 	flashSale := []Flash{}

	// 	err = json.Unmarshal([]byte(fSaleData), &flashSale)
	// 	Expect(err).ToNot(HaveOccurred())

	// 	_, err = dbFlash.UpdateFlashSale(flashSale)
	// 	Expect(err).To(HaveOccurred())
	// })

})
