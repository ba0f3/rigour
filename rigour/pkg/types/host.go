package types

import (
	"time"
)

type Host struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	IP        string    `json:"ip" bson:"ip"`
	IPInt     uint64    `json:"ip_int" bson:"ip_int"`
	ASN       *ASNInfo  `json:"asn,omitempty" bson:"asn,omitempty"`
	Location  *Location `json:"location,omitempty" bson:"location,omitempty"`
	FirstSeen time.Time `json:"first_seen" bson:"first_seen"`
	LastSeen  time.Time `json:"last_seen" bson:"last_seen"`
	Services  []Service `json:"services,omitempty" bson:"services,omitempty"`
	Labels    []string  `json:"labels,omitempty" bson:"labels,omitempty"`
}

type Location struct {
	Coordinates [2]float64 `json:"coordinates" bson:"coordinates"`                                  // [longitude, latitude]
	City        string     `json:"city,omitempty" bson:"city,omitempty"`                               // City name
	Timezone    string     `json:"timezone,omitempty" bson:"timezone,omitempty"`                           // IANA timezone identifier
	CountryCode string     `json:"country_code,omitempty" bson:"country_code,omitempty"` // ISO 3166-1 alpha-2 country code
	CountryName string     `json:"country_name,omitempty" bson:"country_name,omitempty"` // Country name
}

type ASNInfo struct {
	Number              uint32 `json:"number" bson:"number"`                                               // ASN number
	Organization        string `json:"organization,omitempty" bson:"organization,omitempty"`               // ISP/Organization name
	IsSatelliteProvider bool   `json:"is_satellite_provider,omitempty" bson:"is_satellite_provider"` // Whether the ASN is a satellite internet provider
}
