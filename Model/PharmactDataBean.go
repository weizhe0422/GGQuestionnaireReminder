package Model

type PharmacyInfoCollection struct {
	Type     string         `json:"type"`
	Features []PharmacyInfo `json:"features"`
}

type PharmacyInfo struct {
	Type       string           `json:"type"`
	Properties PharmacyProps    `json:"properties"`
	Geometry   PharmacyGeometry `json:"geometry"`
}

type PharmacyProps struct {
	Id              string `json:"id"`
	Name            string `json:"name"`
	Phone           string `json:"phone"`
	Address         string `json:"address"`
	Mask_adult      int    `json:"mask_adult"`
	Mask_child      int    `json:"mask_child"`
	Updated         string `json:"updated"`
	Available       string `json:"available"`
	Note            string `json:"note"`
	Custom_note     string `json:"custom_note"`
	Website         string `json:"website"`
	County          string `json:"county"`
	Town            string `json:"town"`
	Cunli           string `json:"cunli"`
	Service_periods string `json:"service_periods"`
}

type PharmacyGeometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}
