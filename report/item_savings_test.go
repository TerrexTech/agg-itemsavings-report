package report

import (
	ctx "context"
	"encoding/json"
	"log"
	"testing"
	"time"

	"github.com/TerrexTech/uuuid"

	"github.com/TerrexTech/go-mongoutils/mongo"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBooks(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Books Suite")
}

// newTimeoutContext creates a new WithTimeout context with specified timeout.
func newTimeoutContext(timeout uint32) (ctx.Context, ctx.CancelFunc) {
	return ctx.WithTimeout(
		ctx.Background(),
		time.Duration(timeout)*time.Millisecond,
	)
}

var _ = Describe("Mongo service test", func() {
	var (
		// jsonString string
		mgTable *mongo.Collection
		client  *mongo.Client

		item1 WasteItem
		item2 WasteItem
	)

	dropTestDatabase := func() {
		var err error
		client, err = mongo.NewClient(mongo.ClientConfig{
			Hosts:               []string{"mongo:27017"},
			Username:            "root",
			Password:            "root",
			TimeoutMilliseconds: 5000,
		})
		Expect(err).ToNot(HaveOccurred())

		dbCtx, dbCancel := newTimeoutContext(5000)
		err = client.Database("rns_test").Drop(dbCtx)
		dbCancel()
		Expect(err).ToNot(HaveOccurred())

		err = client.Disconnect()
		Expect(err).ToNot(HaveOccurred())
	}

	createTestDatabase := func(collectionName string, schema interface{}) {
		var err error
		client, err = mongo.NewClient(mongo.ClientConfig{
			Hosts:               []string{"mongo:27017"},
			Username:            "root",
			Password:            "root",
			TimeoutMilliseconds: 5000,
		})
		Expect(err).ToNot(HaveOccurred())

		conn := &mongo.ConnectionConfig{
			Client:  client,
			Timeout: 1000,
		}

		mgTable, err = mongo.EnsureCollection(&mongo.Collection{
			Connection:   conn,
			Name:         collectionName,
			Database:     "rns_test",
			SchemaStruct: schema,
		})
		Expect(err).ToNot(HaveOccurred())
	}

	BeforeEach(func() {

		createTestDatabase("mtest", &WasteItem{})
		itemID, err := uuuid.NewV4()
		Expect(err).ToNot(HaveOccurred())
		wasteID, err := uuuid.NewV4()
		Expect(err).ToNot(HaveOccurred())
		item1 = WasteItem{
			ItemID:      itemID,
			WasteID:     wasteID,
			SKU:         "test-sku1",
			Name:        "test-name1",
			Lot:         "test-lot1",
			Weight:      101,
			TotalWeight: 120,
			Timestamp:   10,
		}
		_, err = mgTable.InsertOne(item1)
		Expect(err).ToNot(HaveOccurred())

		itemID, err = uuuid.NewV4()
		Expect(err).ToNot(HaveOccurred())
		wasteID, err = uuuid.NewV4()
		Expect(err).ToNot(HaveOccurred())
		item2 = WasteItem{
			ItemID:      itemID,
			WasteID:     wasteID,
			SKU:         "test-sku2",
			Name:        "test-name2",
			Lot:         "test-lot1",
			Weight:      105,
			TotalWeight: 140,
			Timestamp:   20,
		}
		_, err = mgTable.InsertOne(item2)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		dropTestDatabase()
		err := client.Disconnect()
		Expect(err).ToNot(HaveOccurred())
	})

	It("Confirm inserted ItemID field", func() {
		var findResults []interface{}
		findResults, err := mgTable.Find(map[string]interface{}{
			"itemID": map[string]interface{}{
				"$eq": item1.ItemID,
			},
		})
		Expect(err).ToNot(HaveOccurred())

		for _, v := range findResults {
			m, assertOK := v.(map[string]interface{})
			Expect(assertOK).To(BeTrue())
			log.Println(m)
			Expect(m["itemID"]).To(Equal(item1.ItemID.String()))
			Expect(m["itemID"]).To(Equal(item2.ItemID.String()))
		}
	})

	It("Confirm inserted WasteID field", func() {
		var findResults []interface{}
		findResults, err := mgTable.Find(map[string]interface{}{
			"wasteID": map[string]interface{}{
				"$eq": item1.WasteID,
			},
		})
		Expect(err).ToNot(HaveOccurred())

		for _, v := range findResults {
			m, assertOK := v.(map[string]interface{})
			Expect(assertOK).To(BeTrue())
			log.Println(m["avg_waste"])
			Expect(m["wasteID"]).To(Equal(item1.WasteID.String()))
		}
	})

	It("Check avgSold for inserted records", func() {
		searchParameters := []byte(`{"timestamp":{"$gt":9},"timestamp":{"$lt":21}}`)

		x := WasteItemParams{}
		err := json.Unmarshal(searchParameters, &x)
		Expect(err).ToNot(HaveOccurred())

		avgWasteReport, err := ItemWasteReport(x, mgTable)
		Expect(err).ToNot(HaveOccurred())

		log.Println(avgWasteReport, "*******************")

		for _, v := range avgWasteReport {
			m, assertOK := v.(map[string]interface{})
			Expect(assertOK).To(BeTrue())
			log.Println(m["_id"])
			log.Println(m)

			avgSold, assertOK := m["avg_waste"].(float64)
			Expect(assertOK).To(BeTrue())
			log.Println(avgSold)
			if avgSold == item1.Weight {
				Expect(avgSold).To(Equal(item1.Weight))
				log.Println(item1.Weight)
			} else {
				Expect(avgSold).To(Equal(item2.Weight))
				log.Println(item2.Weight)
			}

		}
	})

	It("Error when timestamp is empty", func() {
		searchParameters := []byte(`{"timestamp":{"$gt":0},"timestamp":{"$lt":0}}`)

		x := WasteItemParams{}
		err := json.Unmarshal(searchParameters, &x)
		Expect(err).ToNot(HaveOccurred())

		_, err = ItemWasteReport(x, mgTable)
		Expect(err).To(HaveOccurred())
	})

	It("Error when $gt timestamp is empty", func() {
		searchParameters := []byte(`{"timestamp":{"$gt":0},"timestamp":{"$lt":10}}`)

		x := WasteItemParams{}
		err := json.Unmarshal(searchParameters, &x)
		Expect(err).ToNot(HaveOccurred())

		_, err = ItemWasteReport(x, mgTable)
		Expect(err).To(HaveOccurred())
	})

	It("Error when $lt timestamp is empty", func() {
		searchParameters := []byte(`{"timestamp":{"$gt":10},"timestamp":{"$lt":0}}`)

		x := WasteItemParams{}
		err := json.Unmarshal(searchParameters, &x)
		Expect(err).ToNot(HaveOccurred())

		_, err = ItemWasteReport(x, mgTable)
		Expect(err).To(HaveOccurred())
	})

	It("Insert into report the avgSold results", func() {
		searchParameters := []byte(`{"timestamp":{"$gt":9},"timestamp":{"$lt":21}}`)

		wasteItemParams := WasteItemParams{}
		err := json.Unmarshal(searchParameters, &wasteItemParams)
		Expect(err).ToNot(HaveOccurred())

		avgWasteReport, err := ItemWasteReport(wasteItemParams, mgTable)
		Expect(err).ToNot(HaveOccurred())

		var reportAgg []ReportResult

		for _, v := range avgWasteReport {
			m, assertOK := v.(map[string]interface{})
			Expect(assertOK).To(BeTrue())

			avgSold, assertOK := m["avg_waste"].(float64)
			Expect(assertOK).To(BeTrue())
			log.Println(avgSold)
			if avgSold == item1.Weight {
				Expect(avgSold).To(Equal(item1.Weight))
				log.Println(item1.Weight)
			} else {
				Expect(avgSold).To(Equal(item2.Weight))
				log.Println(item2.Weight)
			}

			log.Println("$$$$$$$$$$$$$$$$$$", avgWasteReport)

			groupByFields := m["_id"]
			mapInGroupBy := groupByFields.(map[string]interface{})
			sku := mapInGroupBy["sku"].(string)
			name := mapInGroupBy["name"].(string)

			//Generate value for previous year
			currWasteWeight := m["avg_waste"].(float64)
			prevWasteWeight := currWasteWeight * generateRandomFloat(0.1, 2.8)

			amWasteCurrRandPrice := generateRandomFloat(0.5, 5.9)
			amWasteCurr := currWasteWeight * amWasteCurrRandPrice

			log.Println(generateRandomFloat(0.1, amWasteCurrRandPrice))
			amWastePrev := prevWasteWeight * generateRandomFloat(0.1, amWasteCurrRandPrice)
			savingsPercent := ((amWastePrev - amWasteCurr) / amWasteCurr) * 100

			reportAgg = []ReportResult{
				ReportResult{
					SKU:             sku,
					Name:            name,
					WasteWeight:     currWasteWeight,
					PrevWasteWeight: prevWasteWeight,
					AmWastePrev:     amWastePrev,
					AmWasteCurr:     amWasteCurr,
					SavingsPercent:  savingsPercent,
				},
			}
		}

		log.Println(reportAgg)

		createTestDatabase("reportTest", &WasteReport{})
		reportID, err := uuuid.NewV4()
		Expect(err).ToNot(HaveOccurred())

		reportGen := WasteReport{
			ReportID:     reportID,
			SearchQuery:  wasteItemParams,
			ReportResult: reportAgg,
		}

		_, err = CreateReport(reportGen, mgTable)
		Expect(err).ToNot(HaveOccurred())

		var findResults []interface{}
		findResults, err = mgTable.Find(map[string]interface{}{
			"reportID": map[string]interface{}{
				"$eq": reportID,
			},
		})
		Expect(err).ToNot(HaveOccurred())

		for _, v := range findResults {
			log.Println(v.(*WasteReport))

			m := v.(*WasteReport)

			log.Println(m.ReportID.String())
			Expect(m.ReportID.String()).To(Equal(reportID.String()))
		}
	})

})
