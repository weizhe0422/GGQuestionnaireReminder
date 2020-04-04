package Controller

import (
	"encoding/json"
	"fmt"
	"github.com/weizhe0422/GGQuestionnaireReminder/Model"
	"io/ioutil"
	"math"
	"net/http"
)

type PharmacyInfo struct {
	URL string
}

func (p *PharmacyInfo) GetPharmacyResp() ([]byte, error){
	resp, err := http.Get(p.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to get information: %v", err)
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parsing response body: %v", err)
	}
	return content, nil
}

func (p *PharmacyInfo) GetPharmacyInfoList(pharmacyResp []byte) ([]Model.PharmacyInfo, error){
	var pharmacyInfos Model.PharmacyInfoCollection
	if err := json.Unmarshal(pharmacyResp, &pharmacyInfos); err != nil {
		return nil, err
	}
	return pharmacyInfos.Features, nil
}

func rad (x float64) float64{
	return x * math.Pi/180
}
func Haversine (lat1 float64, long1 float64, lat2 float64, long2 float64) float64{
	R := 6371
	dLat := rad(lat2-lat1)
	dLong := rad(long2-long1)
	a := math.Sin(dLat/2) * math.Sin(dLat/2) + math.Cos(rad(lat1)) * math.Cos(rad(lat2)) * math.Sin(dLong/2) * math.Sin(dLong/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return float64(R) * c
}

func (p *PharmacyInfo) GetLastShortDistancePharmacy(allPharmacyList []Model.PharmacyInfo, home_info [2]float64 )  ([5]Model.PharmacyInfo, error) {
	var lastShortDistance [5]Model.PharmacyInfo
	tmpDist := [5]float64{math.MaxFloat64,math.MaxFloat64,math.MaxFloat64,math.MaxFloat64,math.MaxFloat64}

	for _, pharmacy := range allPharmacyList{
		distResult := Haversine(pharmacy.Geometry.Coordinates[1], pharmacy.Geometry.Coordinates[0],home_info[0],home_info[1])
		for idx, value := range tmpDist{
			if distResult < value {
				tmpDist[idx] = distResult
				lastShortDistance[idx] = pharmacy
				break
			}
		}
	}
	return lastShortDistance,nil
}