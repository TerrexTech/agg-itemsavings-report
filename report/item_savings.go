package report

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/TerrexTech/go-mongoutils/mongo"
	"github.com/mongodb/mongo-go-driver/bson"
	mgo "github.com/mongodb/mongo-go-driver/mongo"
	"github.com/pkg/errors"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randomFloat(min, max float64) float64 {
	return rand.Float64() * (max - min)
}

func generateRandomFloat(num1, num2 float64) float64 {
	// rand.Seed(time.Now().Unix())
	return randomFloat(num1, num2)
}

func ItemWasteReport(aggParams WasteItemParams, itemWasteColl *mongo.Collection) ([]interface{}, error) {

	if aggParams.Timestamp.Lt == 0 || aggParams.Timestamp.Gt == 0 {
		err := errors.New("Missing timestamp value")
		log.Println(err)
		return nil, err
	}
	input, err := json.Marshal(aggParams)
	if err != nil {
		err = errors.Wrap(err, "Unable to marshal aggParams")
		log.Println(err)
		return nil, err
	}

	log.Println(input)
	log.Println(aggParams)

	pipelineBuilder := fmt.Sprintf(`[
		{
			"$match": %s
		},
		{
			"$group" : {
			"_id" : {"sku" : "$sku","name":"$name"},
			"avg_waste": {
				"$avg": "$weight",
			}
		}
		}
	]`, input)

	// "avg_total": {
	// 	"$avg": "$totalWeight",
	// }

	pipelineAgg, err := bson.ParseExtJSONArray(pipelineBuilder)
	if err != nil {
		err = errors.Wrap(err, "Query: Error in generating pipeline for report")
		log.Println(err)
		return nil, err
	}

	findResult, err := itemWasteColl.Aggregate(pipelineAgg)
	if err != nil {
		err = errors.Wrap(err, "Query: Error in getting aggregate results ")
		log.Println(err)
		return nil, err
	}
	return findResult, nil
}

func SavingsWasteWeight(avgWasteReport []interface{}) []ReportResult {
	var reportAgg []ReportResult

	for _, v := range avgWasteReport {
		m, assertOK := v.(map[string]interface{})
		if !assertOK {
			err := errors.New("Error getting results ")
			log.Println(err)
		}

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

		reportAgg = append(reportAgg, ReportResult{
			SKU:             sku,
			Name:            name,
			WasteWeight:     currWasteWeight,
			PrevWasteWeight: prevWasteWeight,
			AmWastePrev:     amWastePrev,
			AmWasteCurr:     amWasteCurr,
			SavingsPercent:  savingsPercent,
		})

		// reportAgg = []ReportResult{
		// 	ReportResult{
		// 		SKU:             sku,
		// 		Name:            name,
		// 		WasteWeight:     currWasteWeight,
		// 		PrevWasteWeight: prevWasteWeight,
		// 		AmWastePrev:     amWastePrev,
		// 		AmWasteCurr:     amWasteCurr,
		// 		SavingsPercent:  savingsPercent,
		// 	},
		// }
	}
	return reportAgg
}

func CreateReport(reportGen WasteReport, reportColl *mongo.Collection) (*mgo.InsertOneResult, error) {
	insertRep, err := reportColl.InsertOne(reportGen)
	if err != nil {
		err = errors.Wrap(err, "Query: Error in generating report ")
		log.Println(err)
		return nil, err
	}

	return insertRep, nil
}
