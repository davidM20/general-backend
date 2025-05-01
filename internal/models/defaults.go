package models

// GetDefaultNationalities returns the predefined list of nationalities.
func GetDefaultNationalities() []Nationality {
	return []Nationality{
		{CountryName: "Argentina", IsoCode: "AR", DocIdFormat: `^\d{7,8}$`},
		{CountryName: "Australia", IsoCode: "AU", DocIdFormat: `^\d{8}$`},
		{CountryName: "Austria", IsoCode: "AT", DocIdFormat: `^\d{9}$`},
		{CountryName: "Belgium", IsoCode: "BE", DocIdFormat: `^\d{11}$`},
		{CountryName: "Brazil", IsoCode: "BR", DocIdFormat: `^\d{11}$`},
		{CountryName: "Canada", IsoCode: "CA", DocIdFormat: `^\d{9}$`},
		{CountryName: "Chile", IsoCode: "CL", DocIdFormat: `^\d{1,2}\.\d{3}\.\d{3}-[\dkK]$`},
		{CountryName: "China", IsoCode: "CN", DocIdFormat: `^\d{15}$|^\d{18}$`},
		{CountryName: "Colombia", IsoCode: "CO", DocIdFormat: `^\d{6,10}$`},
		{CountryName: "Costa Rica", IsoCode: "CR", DocIdFormat: `^\d{9}$`},
		{CountryName: "Cuba", IsoCode: "CU", DocIdFormat: `^\d{11}$`},
		{CountryName: "Denmark", IsoCode: "DK", DocIdFormat: `^\d{10}$`},
		{CountryName: "Dominican Republic", IsoCode: "DO", DocIdFormat: `^\d{11}$`},
		{CountryName: "Ecuador", IsoCode: "EC", DocIdFormat: `^\d{10}$`},
		{CountryName: "Egypt", IsoCode: "EG", DocIdFormat: `^\d{14}$`},
		{CountryName: "El Salvador", IsoCode: "SV", DocIdFormat: `^\d{8}$`},
		{CountryName: "Finland", IsoCode: "FI", DocIdFormat: `^\d{6}-\d{4}$`},
		{CountryName: "France", IsoCode: "FR", DocIdFormat: `^\d{15}$`},
		{CountryName: "Germany", IsoCode: "DE", DocIdFormat: `^\d{9}$`},
		{CountryName: "Greece", IsoCode: "GR", DocIdFormat: `^[A-Z]{1}\d{7}$`},
		{CountryName: "Guatemala", IsoCode: "GT", DocIdFormat: `^\d{13}$`},
		{CountryName: "Honduras", IsoCode: "HN", DocIdFormat: `^\d{13}$`},
		{CountryName: "Hungary", IsoCode: "HU", DocIdFormat: `^\d{9}$`},
		{CountryName: "Iceland", IsoCode: "IS", DocIdFormat: `^\d{10}$`},
		{CountryName: "India", IsoCode: "IN", DocIdFormat: `^\d{12}$`},
		{CountryName: "Indonesia", IsoCode: "ID", DocIdFormat: `^\d{16}$`},
		{CountryName: "Ireland", IsoCode: "IE", DocIdFormat: `^\d{9}$`},
		{CountryName: "Israel", IsoCode: "IL", DocIdFormat: `^\d{9}$`},
		{CountryName: "Italy", IsoCode: "IT", DocIdFormat: `^[A-Z]{2}\d{7}[A-Z]{1}$`},
		{CountryName: "Jamaica", IsoCode: "JM", DocIdFormat: `^\d{9}$`},
		{CountryName: "Japan", IsoCode: "JP", DocIdFormat: `^\d{12}$`},
		{CountryName: "Kenya", IsoCode: "KE", DocIdFormat: `^\d{8}$`},
		{CountryName: "South Korea", IsoCode: "KR", DocIdFormat: `^\d{13}$`},
		{CountryName: "Lebanon", IsoCode: "LB", DocIdFormat: `^\d{8}$`},
		{CountryName: "Malaysia", IsoCode: "MY", DocIdFormat: `^\d{12}$`},
		{CountryName: "Mexico", IsoCode: "MX", DocIdFormat: `^[A-Z]{4}\d{6}[A-Z]{6}\d{2}$`},
		{CountryName: "Morocco", IsoCode: "MA", DocIdFormat: `^[A-Z]{1}\d{7}$`},
		{CountryName: "Netherlands", IsoCode: "NL", DocIdFormat: `^\d{9}$`},
		{CountryName: "New Zealand", IsoCode: "NZ", DocIdFormat: `^\d{8}$`},
		{CountryName: "Nicaragua", IsoCode: "NI", DocIdFormat: `^\d{14}$`},
		{CountryName: "Norway", IsoCode: "NO", DocIdFormat: `^\d{11}$`},
		{CountryName: "Pakistan", IsoCode: "PK", DocIdFormat: `^\d{13}$`},
		{CountryName: "Panama", IsoCode: "PA", DocIdFormat: `^\d{8}$`},
		{CountryName: "Paraguay", IsoCode: "PY", DocIdFormat: `^\d{7,8}$`},
		{CountryName: "Peru", IsoCode: "PE", DocIdFormat: `^\d{8}$`},
		{CountryName: "Philippines", IsoCode: "PH", DocIdFormat: `^\d{12}$`},
		{CountryName: "Poland", IsoCode: "PL", DocIdFormat: `^\d{11}$`},
		{CountryName: "Portugal", IsoCode: "PT", DocIdFormat: `^\d{9}$`},
		{CountryName: "Puerto Rico", IsoCode: "PR", DocIdFormat: `^\d{9}$`},
		{CountryName: "Romania", IsoCode: "RO", DocIdFormat: `^\d{13}$`},
		{CountryName: "Russia", IsoCode: "RU", DocIdFormat: `^\d{10,12}$`},
		{CountryName: "Saudi Arabia", IsoCode: "SA", DocIdFormat: `^\d{10}$`},
		{CountryName: "Singapore", IsoCode: "SG", DocIdFormat: `^[A-Z]{1}\d{7}[A-Z]{1}$`},
		{CountryName: "South Africa", IsoCode: "ZA", DocIdFormat: `^\d{13}$`},
		{CountryName: "Spain", IsoCode: "ES", DocIdFormat: `^[A-Z0-9]{8,9}$`},
		{CountryName: "Sri Lanka", IsoCode: "LK", DocIdFormat: `^\d{9}[vV]$`},
		{CountryName: "Sweden", IsoCode: "SE", DocIdFormat: `^\d{12}$`},
		{CountryName: "Switzerland", IsoCode: "CH", DocIdFormat: `^\d{10}$`},
		{CountryName: "Taiwan", IsoCode: "TW", DocIdFormat: `^[A-Z]{1}\d{9}$`},
		{CountryName: "Thailand", IsoCode: "TH", DocIdFormat: `^\d{13}$`},
		{CountryName: "Trinidad and Tobago", IsoCode: "TT", DocIdFormat: `^\d{10}$`},
		{CountryName: "Turkey", IsoCode: "TR", DocIdFormat: `^\d{11}$`},
		{CountryName: "Ukraine", IsoCode: "UA", DocIdFormat: `^\d{9}$`},
		{CountryName: "United Kingdom", IsoCode: "GB", DocIdFormat: `^\d{10}$`},
		{CountryName: "United States", IsoCode: "US", DocIdFormat: `^\d{9}$`},
		{CountryName: "Uruguay", IsoCode: "UY", DocIdFormat: `^\d{7,8}$`},
		{CountryName: "Venezuela", IsoCode: "VE", DocIdFormat: `^\d{6,8}$`},
		{CountryName: "Vietnam", IsoCode: "VN", DocIdFormat: `^\d{9}$`},
		{CountryName: "Zimbabwe", IsoCode: "ZW", DocIdFormat: `^\d{9}$`},
	}
}

// GetDefaultStatusAuthorized returns the predefined list of authorization statuses.
func GetDefaultStatusAuthorized() []StatusAuthorized {
	return []StatusAuthorized{
		{Name: "Active", Id: 1},
		{Name: "Blocked", Id: 2},
		{Name: "Suspended", Id: 3},
		{Name: "Closed", Id: 4},
		{Name: "Pending Verification", Id: 5},
		{Name: "Under Review", Id: 6},
	}
}

// GetDefaultTokensType returns the predefined list of token types.
func GetDefaultTokensType() []Token {
	return []Token{
		{TokenType: "Session", Id: 1},
		{TokenType: "Verification", Id: 2},
		{TokenType: "Access Key", Id: 3},
		{TokenType: "Password Reset", Id: 4},
		{TokenType: "API Key", Id: 5},
		{TokenType: "OAuth", Id: 6},
	}
}

// GetDefaultRoles returns the predefined list of user roles.
func GetDefaultRoles() []Role {
	return []Role{
		{Name: "estudiante-pregrado", Id: 1},
		{Name: "egresado", Id: 2},
		{Name: "moderator", Id: 3},
		{Name: "invitado", Id: 4},
		{Name: "profesor", Id: 5},
		{Name: "personal", Id: 6},
		{Name: "admin", Id: 7},
		{Name: "superadmin", Id: 8},
		{Name: "Empresa", Id: 9},
		{Name: "estudiante-postgrado", Id: 10},
	}
}

// GetDefaultUniversities returns the predefined list of universities.
func GetDefaultUniversities() []University {
	// We assign a temporary ID here only for the default degree mapping below.
	// The actual ID will be assigned by AUTO_INCREMENT in the DB.
	return []University{
		{
			Id:     1, // Temporary ID for mapping
			Name:   "Santa Maria",
			Campus: "Florencia",
		},
	}
}

// GetDefaultDegrees returns the predefined list of degrees.
func GetDefaultDegrees() []Degree {
	// UniversityId here refers to the temporary ID assigned in GetDefaultUniversities.
	return []Degree{
		{
			DegreeName:   "Ingenieria en Sitemas",
			UniversityId: 1, // Corresponds to Santa Maria
			Descriptions: "Carrera de ingenieria que se dedica a la informatica",
			Code:         "ING-SIS",
		},
	}
}

// GetDefaultTypeMessages returns the predefined list of message types.
func GetDefaultTypeMessages() []TypeMessage {
	return []TypeMessage{
		{
			Id:          1,
			Name:        "Text",
			Description: "Texto plano",
		},
		{
			Id:          2,
			Name:        "Audio",
			Description: "Audio",
		},
		{
			Id:          3,
			Name:        "Image",
			Description: "Image",
		},
		{
			Id:          4,
			Name:        "Video",
			Description: "Video",
		},
		{
			Id:          5,
			Name:        "Document",
			Description: "Documento",
		},
		{
			Id:          6,
			Name:        "Location",
			Description: "Ubicación",
		},
		{
			Id:          7,
			Name:        "Contact",
			Description: "Contacto",
		},
		{
			Id:          8,
			Name:        "Sticker",
			Description: "Sticker",
		}, {
			Id:          9,
			Name:        "Gif",
			Description: "Gif",
		},
	}
}
