package models

import "time"

// Token defines the structure for the Token table.
type Token struct {
	Id        int    `json:"id" db:"Id"`
	TokenType string `json:"token_type" db:"TokenType"`
}

// Category defines the structure for the Category table.
type Category struct {
	CategoryId  int64  `json:"category_id" db:"CategoryId"`
	Name        string `json:"name" db:"Name"`
	Description string `json:"description" db:"Description"`
}

// Interest defines the structure for the Interest table.
type Interest struct {
	InterestId      int64  `json:"interest_id" db:"InterestId"`
	CategoryId      int64  `json:"category_id" db:"CategoryId"`
	Description     string `json:"description" db:"Description"`
	ExperienceLevel string `json:"experience_level" db:"ExperienceLevel"`
}

// TypeMessage defines the structure for the TypeMessage table.
type TypeMessage struct {
	Id          int64  `json:"id" db:"Id"`
	Name        string `json:"name" db:"Name"`
	Description string `json:"description" db:"Description"`
}

// Nationality defines the structure for the Nationality table.
type Nationality struct {
	Id          int    `json:"id" db:"Id"`
	CountryName string `json:"country_name" db:"CountryName"`
	IsoCode     string `json:"iso_code" db:"IsoCode"`
	DocIdFormat string `json:"doc_id_format" db:"DocIdFormat"`
}

// StatusAuthorized defines the structure for the StatusAuthorized table.
type StatusAuthorized struct {
	Id   int    `json:"id" db:"Id"`
	Name string `json:"name" db:"Name"`
}

// University defines the structure for the University table.
type University struct {
	Id     int64  `json:"id" db:"Id"`
	Name   string `json:"name" db:"Name"`
	Campus string `json:"campus" db:"Campus"`
}

// Degree defines the structure for the Degree table.
type Degree struct {
	Id           int64  `json:"id" db:"Id"`
	DegreeName   string `json:"degree_name" db:"DegreeName"`
	Descriptions string `json:"descriptions" db:"Descriptions"`
	Code         string `json:"code" db:"Code"`
	UniversityId int64  `json:"university_id" db:"UniversityId"`
}

// Role defines the structure for the Role table.
type Role struct {
	Id   int    `json:"id" db:"Id"`
	Name string `json:"name" db:"Name"`
}

// User defines the structure for the User table.
type User struct {
	Id                 int64      `json:"id" db:"Id"`
	FirstName          string     `json:"first_name" db:"FirstName"`
	LastName           string     `json:"last_name" db:"LastName"`
	UserName           string     `json:"user_name" db:"UserName"`
	Password           string     `json:"-" db:"Password"` // Exclude password from JSON responses
	Email              string     `json:"email" db:"Email"`
	Phone              string     `json:"phone" db:"Phone"`
	Sex                string     `json:"sex" db:"Sex"`
	DocId              string     `json:"doc_id" db:"DocId"`
	NationalityId      int        `json:"nationality_id" db:"NationalityId"`
	Birthdate          *time.Time `json:"birthdate" db:"Birthdate"` // Use pointer for nullable date
	Picture            string     `json:"picture" db:"Picture"`
	DegreeId           int64      `json:"degree_id" db:"DegreeId"`
	UniversityId       int64      `json:"university_id" db:"UniversityId"`
	RoleId             int        `json:"role_id" db:"RoleId"`
	StatusAuthorizedId int        `json:"status_authorized_id" db:"StatusAuthorizedId"`
	Summary            string     `json:"summary" db:"Summary"`
	Address            string     `json:"address" db:"Address"`
	Github             string     `json:"github" db:"Github"`
	Linkedin           string     `json:"linkedin" db:"Linkedin"`
}

// Online defines the structure for the Online table.
type Online struct {
	UserOnlineId int64     `json:"user_online_id" db:"UserOnlineId"`
	CreateAt     time.Time `json:"create_at" db:"CreateAt"`
	Status       bool      `json:"status" db:"Status"` // Changed TINYINT(1) to bool
}

// Contact defines the structure for the Contact table.
type Contact struct {
	ContactId int64  `json:"contact_id" db:"ContactId"`
	User1Id   int64  `json:"user1_id" db:"User1Id"`
	User2Id   int64  `json:"user2_id" db:"User2Id"`
	Status    string `json:"status" db:"Status"`
	ChatId    string `json:"chat_id" db:"ChatId"`
}

// Multimedia defines the structure for the Multimedia table.
type Multimedia struct {
	Id        string    `json:"id" db:"Id"`
	Type      string    `json:"type" db:"Type"` // e.g., image, video, audio
	Ratio     float32   `json:"ratio" db:"Ratio"`
	UserId    int64     `json:"user_id" db:"UserId"`
	FileName  string    `json:"file_name" db:"FileName"` // Original or unique generated name? Assume unique.
	CreateAt  time.Time `json:"create_at" db:"CreateAt"`
	ContentId string    `json:"content_id" db:"ContentId"` // ID for the content in storage
	ChatId    string    `json:"chat_id" db:"ChatId"`
}

// Session defines the structure for the Session table.
type Session struct {
	Id      int64  `json:"id" db:"Id"`
	UserId  int64  `json:"user_id" db:"UserId"`
	Tk      string `json:"tk" db:"Tk"` // JWT token
	Ip      string `json:"ip" db:"Ip"`
	RoleId  int    `json:"role_id" db:"RoleId"`
	TokenId int    `json:"token_id" db:"TokenId"` // Refers to Token.Id
}

// Message defines the structure for the Message table.
type Message struct {
	Id            string    `json:"id" db:"Id"`
	TypeMessageId int64     `json:"type_message_id" db:"TypeMessageId"`
	Text          string    `json:"text" db:"Text"`
	MediaId       string    `json:"media_id" db:"MediaId"` // Refers to Multimedia.Id
	Date          time.Time `json:"date" db:"Date"`
	StatusMessage int       `json:"status_message" db:"StatusMessage"` // e.g., sent, delivered, read
	UserId        int64     `json:"user_id" db:"UserId"`
	ChatId        string    `json:"chat_id" db:"ChatId"`
	ResponseTo    string    `json:"response_to" db:"ResponseTo"` // Refers to Message.Id
}

// Education defines the structure for the Education table.
type Education struct {
	Id             int64      `json:"id" db:"Id"`
	PersonId       int64      `json:"person_id" db:"PersonId"`
	Institution    string     `json:"institution" db:"Institution"`
	Degree         string     `json:"degree" db:"Degree"`
	Campus         string     `json:"campus" db:"Campus"`
	GraduationDate *time.Time `json:"graduation_date" db:"GraduationDate"`
	CountryId      int64      `json:"country_id" db:"CountryId"` // Should reference Nationality? Schema ambiguous
}

// WorkExperience defines the structure for the WorkExperience table.
type WorkExperience struct {
	Id          int64      `json:"id" db:"Id"`
	PersonId    int64      `json:"person_id" db:"PersonId"`
	Company     string     `json:"company" db:"Company"`
	Position    string     `json:"position" db:"Position"`
	StartDate   *time.Time `json:"start_date" db:"StartDate"`
	EndDate     *time.Time `json:"end_date" db:"EndDate"` // Nullable if currently employed
	Description string     `json:"description" db:"Description"`
	CountryId   int64      `json:"country_id" db:"CountryId"` // Should reference Nationality? Schema ambiguous
}

// Certifications defines the structure for the Certifications table.
type Certifications struct {
	Id            int64      `json:"id" db:"Id"`
	PersonId      int64      `json:"person_id" db:"PersonId"`
	Certification string     `json:"certification" db:"Certification"`
	Institution   string     `json:"institution" db:"Institution"`
	DateObtained  *time.Time `json:"date_obtained" db:"DateObtained"`
}

// Skills defines the structure for the Skills table.
type Skills struct {
	Id       int64  `json:"id" db:"Id"`
	PersonId int64  `json:"person_id" db:"PersonId"`
	Skill    string `json:"skill" db:"Skill"`
	Level    string `json:"level" db:"Level"` // e.g., Basic, Intermediate, Advanced
}

// Languages defines the structure for the Languages table.
type Languages struct {
	Id       int64  `json:"id" db:"Id"`
	PersonId int64  `json:"person_id" db:"PersonId"`
	Language string `json:"language" db:"Language"`
	Level    string `json:"level" db:"Level"` // e.g., A1, A2, B1, B2, C1, C2, Native
}

// Project defines the structure for the Project table.
type Project struct {
	Id              int64      `json:"id" db:"Id"`
	PersonID        int64      `json:"person_id" db:"PersonID"`
	Title           string     `json:"title" db:"Title"`
	Role            string     `json:"role" db:"Role"`
	Description     string     `json:"description" db:"Description"`
	Company         string     `json:"company" db:"Company"`
	Document        string     `json:"document" db:"Document"` // Link to document or identifier?
	ProjectStatus   string     `json:"project_status" db:"ProjectStatus"`
	StartDate       *time.Time `json:"start_date" db:"StartDate"`
	ExpectedEndDate *time.Time `json:"expected_end_date" db:"ExpectedEndDate"`
}

// Event defines the structure for the Event (notifications) table.
type Event struct {
	Id          int64     `json:"id" db:"Id"`
	Description string    `json:"description" db:"Description"`
	UserId      int64     `json:"user_id" db:"UserId"`            // User receiving the notification
	OtherUserId int64     `json:"other_user_id" db:"OtherUserId"` // User causing the event (optional)
	ProyectId   int64     `json:"project_id" db:"ProyectId"`      // Related project (optional)
	CreateAt    time.Time `json:"create_at" db:"CreateAt"`
	IsRead      bool      `json:"is_read" db:"IsRead"` // Added field to track read status
}

// Enterprise defines the structure for the Enterprise table.
type Enterprise struct {
	Id          int64  `json:"id" db:"Id"`
	RIF         string `json:"rif" db:"RIF"`
	CompanyName string `json:"company_name" db:"CompanyName"`
	CategoryId  int64  `json:"category_id" db:"CategoryId"`
	Description string `json:"description" db:"Description"`
	Location    string `json:"location" db:"Location"`
	Phone       string `json:"phone" db:"Phone"`
}

// --- Helper Structs ---

// RegistrationStep1 defines the structure for the first step of user registration.
type RegistrationStep1 struct {
	FirstName string `json:"primerNombre"`
	LastName  string `json:"primerApellido"`
	UserName  string `json:"UserName"`
	Email     string `json:"Email"`
	Phone     string `json:"Phone"`
	Password  string `json:"Password"`
}

// RegistrationStep2 defines the structure for the second step of user registration.
type RegistrationStep2 struct {
	DocId         string `json:"DocId"`
	NationalityId int    `json:"NationalityId"`
}

// RegistrationStep3 defines the structure for the third step of user registration.
type RegistrationStep3 struct {
	Sex       string    `json:"Sex"`
	Birthdate time.Time `json:"Birthdate"`
}

// LoginRequest defines the structure for login requests.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse defines the structure for login responses.
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"` // Send back user info (excluding password)
}

// WebSocketMessage defines the generic structure for incoming WebSocket messages.
type WebSocketMessage struct {
	Type    string      `json:"type"` // Corresponds to MessageType* constants
	Payload interface{} `json:"payload"`
}
