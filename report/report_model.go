package report

import (
	"github.com/TerrexTech/uuuid"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/objectid"
	"github.com/pkg/errors"
)

type WasteReport struct {
	ID           objectid.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	ReportID     uuuid.UUID        `bson:"reportID,omitempty" json:"reportID,omitempty"`
	SearchQuery  WasteItemParams   `bson:"searchQuery,omitempty" json:"searchQuery,omitempty"`
	ReportResult []ReportResult    `bson:"reportResult,omitempty" json:"reportResult,omitempty"`
}

type WasteReportBSON struct {
	ID           objectid.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	ReportID     string            `bson:"reportID,omitempty" json:"reportID,omitempty"`
	SearchQuery  WasteItemParams   `bson:"searchQuery,omitempty" json:"searchQuery,omitempty"`
	ReportResult []ReportResult    `bson:"reportResult,omitempty" json:"reportResult,omitempty"`
}

type ReportResult struct {
	SKU             string  `bson:"sku,omitempty" json:"sku,omitempty"`
	Name            string  `bson:"name,omitempty" json:"name,omitempty"`
	PrevWasteWeight float64 `bson:"prevWasteWeight,omitempty" json:"prevWasteWeight,omitempty"`
	WasteWeight     float64 `bson:"wasteWeight,omitempty" json:"wasteWeight,omitempty"`
	// TotalWeight float64 `bson:"totalWeight,omitempty" json:"totalWeight,omitempty"`
	AmWastePrev    float64 `bson:"amWastePrev,omitempty" json:"amWastePrev,omitempty"`
	AmWasteCurr    float64 `bson:"amWasteCurr,omitempty" json:"amWasteCurr,omitempty"`
	SavingsPercent float64 `bson:"savingsPercent,omitempty" json:"savingsPercent,omitempty"`
}

func (s WasteReport) MarshalBSON() ([]byte, error) {
	sm := map[string]interface{}{
		"reportid":     s.ReportID.String(),
		"searchquery":  s.SearchQuery,
		"reportresult": s.ReportResult,
	}
	if s.ID != objectid.NilObjectID {
		sm["_id"] = s.ID
	}
	// Do more stuff

	if s.ReportID != (uuuid.UUID{}) {
		sm["reportID"] = s.ReportID.String()
	}

	return bson.Marshal(sm)
}

func (s *WasteReport) UnmarshalBSON(in []byte) error {
	sb := &WasteReportBSON{}
	err := bson.Unmarshal(in, sb)
	if err != nil {
		err = errors.Wrap(err, "UnmarshalBSON Error")
		return err
	}

	if sb.ID != objectid.NilObjectID {
		s.ID = sb.ID
	}
	reportID, err := uuuid.FromString(sb.ReportID)
	if err != nil {
		err = errors.Wrap(err, "UnmarshalBSON Error: Error parsing SaleID")
	}
	s.ReportID = reportID

	if s.ReportResult == nil {
		s.ReportResult = make([]ReportResult, 0)
	}
	for _, v := range sb.ReportResult {
		s.ReportResult = append(s.ReportResult, ReportResult{
			SKU:             v.SKU,
			Name:            v.Name,
			PrevWasteWeight: v.PrevWasteWeight,
			AmWastePrev:     v.AmWastePrev,
			AmWasteCurr:     v.AmWasteCurr,
			SavingsPercent:  v.SavingsPercent,
		})
	}
	return nil
}
