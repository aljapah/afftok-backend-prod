package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// ============================================
// GEO RULE MODEL
// ============================================

// GeoRuleScopeType represents the scope of a geo rule
type GeoRuleScopeType string

const (
	GeoRuleScopeOffer      GeoRuleScopeType = "offer"
	GeoRuleScopeAdvertiser GeoRuleScopeType = "advertiser"
	GeoRuleScopeGlobal     GeoRuleScopeType = "global"
)

// GeoRuleMode represents the mode of a geo rule
type GeoRuleMode string

const (
	GeoRuleModeAllow GeoRuleMode = "allow"
	GeoRuleModeBlock GeoRuleMode = "block"
)

// GeoRuleStatus represents the status of a geo rule
type GeoRuleStatus string

const (
	GeoRuleStatusActive   GeoRuleStatus = "active"
	GeoRuleStatusDisabled GeoRuleStatus = "disabled"
)

// GeoRule represents a geographic rule for blocking/allowing countries
type GeoRule struct {
	ID        uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	
	// Scope
	ScopeType GeoRuleScopeType `gorm:"size:20;not null;index:idx_geo_rules_scope" json:"scope_type"`
	ScopeID   *uuid.UUID       `gorm:"type:uuid;index:idx_geo_rules_scope_id" json:"scope_id,omitempty"` // nullable for global rules
	
	// Rule configuration
	Mode      GeoRuleMode      `gorm:"size:10;not null;default:'block'" json:"mode"`
	Countries datatypes.JSON   `gorm:"type:jsonb;not null" json:"countries"` // ["US", "KW", "SA"]
	
	// Priority (lower = higher priority)
	Priority  int              `gorm:"default:100;index:idx_geo_rules_priority" json:"priority"`
	
	// Status
	Status    GeoRuleStatus    `gorm:"size:20;default:'active';index:idx_geo_rules_status" json:"status"`
	
	// Metadata
	Name        string         `gorm:"size:100" json:"name,omitempty"`
	Description string         `gorm:"size:500" json:"description,omitempty"`
	
	// Timestamps
	CreatedAt time.Time        `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time        `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	
	// Relations (for preloading)
	Offer      *Offer          `gorm:"foreignKey:ScopeID" json:"offer,omitempty"`
	Advertiser *AfftokUser     `gorm:"foreignKey:ScopeID" json:"advertiser,omitempty"`
}

// TableName returns the table name for GORM
func (GeoRule) TableName() string {
	return "geo_rules"
}

// IsActive returns true if the rule is active
func (r *GeoRule) IsActive() bool {
	return r.Status == GeoRuleStatusActive
}

// GetCountries returns the countries as a string slice
func (r *GeoRule) GetCountries() []string {
	if r.Countries == nil {
		return []string{}
	}
	
	var countries []string
	if err := r.Countries.UnmarshalJSON(r.Countries); err != nil {
		return []string{}
	}
	return countries
}

// ContainsCountry checks if the rule contains a specific country
func (r *GeoRule) ContainsCountry(countryCode string) bool {
	countries := r.GetCountries()
	for _, c := range countries {
		if c == countryCode {
			return true
		}
	}
	return false
}

// ============================================
// GEO RULE DTOs
// ============================================

// CreateGeoRuleRequest represents a request to create a geo rule
type CreateGeoRuleRequest struct {
	ScopeType   string   `json:"scope_type" binding:"required,oneof=offer advertiser global"`
	ScopeID     string   `json:"scope_id,omitempty"`
	Mode        string   `json:"mode" binding:"required,oneof=allow block"`
	Countries   []string `json:"countries" binding:"required,min=1"`
	Priority    int      `json:"priority,omitempty"`
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
}

// UpdateGeoRuleRequest represents a request to update a geo rule
type UpdateGeoRuleRequest struct {
	Mode        *string  `json:"mode,omitempty"`
	Countries   []string `json:"countries,omitempty"`
	Priority    *int     `json:"priority,omitempty"`
	Status      *string  `json:"status,omitempty"`
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
}

// GeoRuleInfo represents geo rule info for API responses
type GeoRuleInfo struct {
	ID          uuid.UUID        `json:"id"`
	ScopeType   GeoRuleScopeType `json:"scope_type"`
	ScopeID     *uuid.UUID       `json:"scope_id,omitempty"`
	ScopeName   string           `json:"scope_name,omitempty"` // Offer/Advertiser name
	Mode        GeoRuleMode      `json:"mode"`
	Countries   []string         `json:"countries"`
	Priority    int              `json:"priority"`
	Status      GeoRuleStatus    `json:"status"`
	Name        string           `json:"name,omitempty"`
	Description string           `json:"description,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// GeoCheckResult represents the result of a geo check
type GeoCheckResult struct {
	Allowed bool       `json:"allowed"`
	Rule    *GeoRule   `json:"rule,omitempty"`
	Reason  string     `json:"reason"` // "blocked_by_rule", "allowed", "no_rule"
}

// ============================================
// ISO 3166-1 ALPHA-2 COUNTRY CODES
// ============================================

// ValidCountryCodes contains all valid ISO 3166-1 alpha-2 country codes
var ValidCountryCodes = map[string]string{
	"AF": "Afghanistan", "AX": "Åland Islands", "AL": "Albania", "DZ": "Algeria",
	"AS": "American Samoa", "AD": "Andorra", "AO": "Angola", "AI": "Anguilla",
	"AQ": "Antarctica", "AG": "Antigua and Barbuda", "AR": "Argentina", "AM": "Armenia",
	"AW": "Aruba", "AU": "Australia", "AT": "Austria", "AZ": "Azerbaijan",
	"BS": "Bahamas", "BH": "Bahrain", "BD": "Bangladesh", "BB": "Barbados",
	"BY": "Belarus", "BE": "Belgium", "BZ": "Belize", "BJ": "Benin",
	"BM": "Bermuda", "BT": "Bhutan", "BO": "Bolivia", "BQ": "Bonaire",
	"BA": "Bosnia and Herzegovina", "BW": "Botswana", "BV": "Bouvet Island", "BR": "Brazil",
	"IO": "British Indian Ocean Territory", "BN": "Brunei", "BG": "Bulgaria", "BF": "Burkina Faso",
	"BI": "Burundi", "CV": "Cabo Verde", "KH": "Cambodia", "CM": "Cameroon",
	"CA": "Canada", "KY": "Cayman Islands", "CF": "Central African Republic", "TD": "Chad",
	"CL": "Chile", "CN": "China", "CX": "Christmas Island", "CC": "Cocos Islands",
	"CO": "Colombia", "KM": "Comoros", "CG": "Congo", "CD": "Congo (DRC)",
	"CK": "Cook Islands", "CR": "Costa Rica", "CI": "Côte d'Ivoire", "HR": "Croatia",
	"CU": "Cuba", "CW": "Curaçao", "CY": "Cyprus", "CZ": "Czechia",
	"DK": "Denmark", "DJ": "Djibouti", "DM": "Dominica", "DO": "Dominican Republic",
	"EC": "Ecuador", "EG": "Egypt", "SV": "El Salvador", "GQ": "Equatorial Guinea",
	"ER": "Eritrea", "EE": "Estonia", "SZ": "Eswatini", "ET": "Ethiopia",
	"FK": "Falkland Islands", "FO": "Faroe Islands", "FJ": "Fiji", "FI": "Finland",
	"FR": "France", "GF": "French Guiana", "PF": "French Polynesia", "TF": "French Southern Territories",
	"GA": "Gabon", "GM": "Gambia", "GE": "Georgia", "DE": "Germany",
	"GH": "Ghana", "GI": "Gibraltar", "GR": "Greece", "GL": "Greenland",
	"GD": "Grenada", "GP": "Guadeloupe", "GU": "Guam", "GT": "Guatemala",
	"GG": "Guernsey", "GN": "Guinea", "GW": "Guinea-Bissau", "GY": "Guyana",
	"HT": "Haiti", "HM": "Heard Island", "VA": "Holy See", "HN": "Honduras",
	"HK": "Hong Kong", "HU": "Hungary", "IS": "Iceland", "IN": "India",
	"ID": "Indonesia", "IR": "Iran", "IQ": "Iraq", "IE": "Ireland",
	"IM": "Isle of Man", "IL": "Israel", "IT": "Italy", "JM": "Jamaica",
	"JP": "Japan", "JE": "Jersey", "JO": "Jordan", "KZ": "Kazakhstan",
	"KE": "Kenya", "KI": "Kiribati", "KP": "North Korea", "KR": "South Korea",
	"KW": "Kuwait", "KG": "Kyrgyzstan", "LA": "Laos", "LV": "Latvia",
	"LB": "Lebanon", "LS": "Lesotho", "LR": "Liberia", "LY": "Libya",
	"LI": "Liechtenstein", "LT": "Lithuania", "LU": "Luxembourg", "MO": "Macao",
	"MG": "Madagascar", "MW": "Malawi", "MY": "Malaysia", "MV": "Maldives",
	"ML": "Mali", "MT": "Malta", "MH": "Marshall Islands", "MQ": "Martinique",
	"MR": "Mauritania", "MU": "Mauritius", "YT": "Mayotte", "MX": "Mexico",
	"FM": "Micronesia", "MD": "Moldova", "MC": "Monaco", "MN": "Mongolia",
	"ME": "Montenegro", "MS": "Montserrat", "MA": "Morocco", "MZ": "Mozambique",
	"MM": "Myanmar", "NA": "Namibia", "NR": "Nauru", "NP": "Nepal",
	"NL": "Netherlands", "NC": "New Caledonia", "NZ": "New Zealand", "NI": "Nicaragua",
	"NE": "Niger", "NG": "Nigeria", "NU": "Niue", "NF": "Norfolk Island",
	"MK": "North Macedonia", "MP": "Northern Mariana Islands", "NO": "Norway", "OM": "Oman",
	"PK": "Pakistan", "PW": "Palau", "PS": "Palestine", "PA": "Panama",
	"PG": "Papua New Guinea", "PY": "Paraguay", "PE": "Peru", "PH": "Philippines",
	"PN": "Pitcairn", "PL": "Poland", "PT": "Portugal", "PR": "Puerto Rico",
	"QA": "Qatar", "RE": "Réunion", "RO": "Romania", "RU": "Russia",
	"RW": "Rwanda", "BL": "Saint Barthélemy", "SH": "Saint Helena", "KN": "Saint Kitts and Nevis",
	"LC": "Saint Lucia", "MF": "Saint Martin", "PM": "Saint Pierre and Miquelon", "VC": "Saint Vincent",
	"WS": "Samoa", "SM": "San Marino", "ST": "São Tomé and Príncipe", "SA": "Saudi Arabia",
	"SN": "Senegal", "RS": "Serbia", "SC": "Seychelles", "SL": "Sierra Leone",
	"SG": "Singapore", "SX": "Sint Maarten", "SK": "Slovakia", "SI": "Slovenia",
	"SB": "Solomon Islands", "SO": "Somalia", "ZA": "South Africa", "GS": "South Georgia",
	"SS": "South Sudan", "ES": "Spain", "LK": "Sri Lanka", "SD": "Sudan",
	"SR": "Suriname", "SJ": "Svalbard", "SE": "Sweden", "CH": "Switzerland",
	"SY": "Syria", "TW": "Taiwan", "TJ": "Tajikistan", "TZ": "Tanzania",
	"TH": "Thailand", "TL": "Timor-Leste", "TG": "Togo", "TK": "Tokelau",
	"TO": "Tonga", "TT": "Trinidad and Tobago", "TN": "Tunisia", "TR": "Turkey",
	"TM": "Turkmenistan", "TC": "Turks and Caicos", "TV": "Tuvalu", "UG": "Uganda",
	"UA": "Ukraine", "AE": "United Arab Emirates", "GB": "United Kingdom", "US": "United States",
	"UM": "U.S. Minor Outlying Islands", "UY": "Uruguay", "UZ": "Uzbekistan", "VU": "Vanuatu",
	"VE": "Venezuela", "VN": "Vietnam", "VG": "British Virgin Islands", "VI": "U.S. Virgin Islands",
	"WF": "Wallis and Futuna", "EH": "Western Sahara", "YE": "Yemen", "ZM": "Zambia",
	"ZW": "Zimbabwe",
}

// IsValidCountryCode checks if a country code is valid
func IsValidCountryCode(code string) bool {
	_, exists := ValidCountryCodes[code]
	return exists
}

// GetCountryName returns the name of a country by code
func GetCountryName(code string) string {
	name, exists := ValidCountryCodes[code]
	if !exists {
		return "Unknown"
	}
	return name
}

// ValidateCountryCodes validates a list of country codes
func ValidateCountryCodes(codes []string) ([]string, []string) {
	valid := make([]string, 0)
	invalid := make([]string, 0)
	
	for _, code := range codes {
		if IsValidCountryCode(code) {
			valid = append(valid, code)
		} else {
			invalid = append(invalid, code)
		}
	}
	
	return valid, invalid
}

