package models

import (
	"database/sql"
	"fmt"
	"time"
)

// REGLAS A SEGUIR
// 1. Si un campo es NULL, se debe manejar con sql.NullString, sql.NullInt64, etc.
// 2. Si un campo es opcional, se debe manejar con *string, *int64, etc.
// 3. Si un campo es un array, se debe manejar con []string, []int64, etc.
// 4. Si un campo es un objeto, se debe manejar con el tipo de objeto definido en el modelo.
// 5. Si un campo es un booleano, se debe manejar con bool.

// NO CAMBIAR EL NOMBRE DE LAS STRUCTS EXISTENTE PARA HACER COMPATIBLE CON UNA NUEVA FUCIONALIDAD
// TAMPOCO CAMBIR LOS TIPOS DE DATOS SI SE VA A USAR ESTE ARCHIVO DEBE SER O PARA LEER O AGREGRA ALGO NUEVO MAS NO CAMBIR LO EXIXTENTE
// LAS FUNCIONES NUEVA SI REQUIREN ALGO SE DEBE DE UTILIZAR LOS MODELS EXISTENTES A MENOS QUE LA TAREA SEA MUY ESPECIFICA PARA UNA FUNCION EN ESPECIFICO O UN CASO ESPECIFICO

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

// User defines the structure for the User table, handling potential NULL values.
type User struct {
	Id                 int64          `json:"id" db:"Id"`
	FirstName          sql.NullString `json:"first_name" db:"FirstName"` // Handle NULL
	LastName           sql.NullString `json:"last_name" db:"LastName"`   // Handle NULL
	UserName           string         `json:"user_name" db:"UserName"`
	Password           string         `json:"-" db:"Password"` // Exclude password from JSON responses
	Email              string         `json:"email" db:"Email"`
	Phone              sql.NullString `json:"phone" db:"Phone"`                  // Handle NULL
	Sex                sql.NullString `json:"sex" db:"Sex"`                      // Handle NULL (Asumiendo que puede ser NULL)
	DocId              sql.NullString `json:"doc_id" db:"DocId"`                 // Handle NULL
	NationalityId      sql.NullInt32  `json:"nationality_id" db:"NationalityId"` // Handle NULL (int es int32)
	NationalityName    sql.NullString `json:"nationality_name,omitempty" db:"NationalityName"`
	Birthdate          sql.NullTime   `json:"birthdate" db:"Birthdate"` // Handle NULL
	Picture            sql.NullString `json:"picture" db:"Picture"`     // Handle NULL
	DegreeId           sql.NullInt64  `json:"degree_id" db:"DegreeId"`  // Handle NULL
	DegreeName         sql.NullString `json:"degree_name,omitempty" db:"DegreeName"`
	UniversityId       sql.NullInt64  `json:"university_id" db:"UniversityId"` // Handle NULL
	UniversityName     sql.NullString `json:"university_name,omitempty" db:"UniversityName"`
	RoleId             int            `json:"role_id" db:"RoleId"` // Asumiendo no NULL
	RoleName           sql.NullString `json:"role_name,omitempty" db:"RoleName"`
	StatusAuthorizedId int            `json:"status_authorized_id" db:"StatusAuthorizedId"` // Asumiendo no NULL
	Summary            sql.NullString `json:"summary" db:"Summary"`                         // Handle NULL
	Address            sql.NullString `json:"address" db:"Address"`                         // Handle NULL
	Github             sql.NullString `json:"github" db:"Github"`                           // Handle NULL
	Linkedin           sql.NullString `json:"linkedin" db:"Linkedin"`                       // Handle NULL
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

// ChatInfoQueryResult es una estructura para contener los resultados de la consulta de lista de chat optimizada.
type ChatInfoQueryResult struct {
	ChatID                string
	OtherUserID           int64
	OtherUserRoleID       int
	OtherUserName         sql.NullString
	OtherFirstName        sql.NullString
	OtherLastName         sql.NullString
	OtherCompanyName      sql.NullString
	OtherPicture          sql.NullString
	LastMessage           sql.NullString
	LastMessageTs         sql.NullTime
	LastMessageFromUserId sql.NullInt64
	UnreadCount           int
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
	Id                  int64          `json:"ID" db:"Id"`
	PersonId            int64          `json:"PersonId" db:"PersonId"`
	Institution         string         `json:"Institution" db:"Institution"`
	Degree              string         `json:"Degree" db:"Degree"`
	Campus              sql.NullString `json:"Campus" db:"Campus"`
	GraduationDate      sql.NullTime   `json:"GraduationDate" db:"GraduationDate"`
	CountryId           sql.NullInt64  `json:"CountryId" db:"CountryId"`
	CountryName         sql.NullString `json:"CountryName,omitempty" db:"CountryName"`
	IsCurrentlyStudying sql.NullBool   `json:"isCurrentlyStudying,omitempty" db:"IsCurrentlyStudying"`
}

// WorkExperience defines the structure for the WorkExperience table.
type WorkExperience struct {
	Id           int64          `json:"ID" db:"Id"`
	PersonId     int64          `json:"PersonId" db:"PersonId"`
	Company      string         `json:"Company" db:"Company"`
	Position     string         `json:"Position" db:"Position"`
	StartDate    sql.NullTime   `json:"StartDate" db:"StartDate"`
	EndDate      sql.NullTime   `json:"EndDate" db:"EndDate"`
	Description  sql.NullString `json:"Description" db:"Description"`
	CountryId    sql.NullInt64  `json:"CountryId" db:"CountryId"`
	CountryName  sql.NullString `json:"CountryName,omitempty" db:"CountryName"`
	IsCurrentJob sql.NullBool   `json:"isCurrentJob,omitempty" db:"IsCurrentJob"`
}

// Certifications defines the structure for the Certifications table.
type Certifications struct {
	Id              int64          `json:"ID" db:"Id"`
	PersonId        int64          `json:"PersonId" db:"PersonId"`
	Certification   string         `json:"Certification" db:"Certification"`
	Institution     string         `json:"Institution" db:"Institution"`
	DateObtained    sql.NullTime   `json:"dateObtained" db:"DateObtained"` // Handle NULL
	Description     sql.NullString `json:"description" db:"Description"`
	Company         sql.NullString `json:"company" db:"Company"`
	Document        sql.NullString `json:"document" db:"Document"`
	ProjectStatus   sql.NullString `json:"project_status" db:"ProjectStatus"`
	StartDate       sql.NullTime   `json:"start_date" db:"StartDate"`              // Handle NULL
	ExpectedEndDate sql.NullTime   `json:"expected_end_date" db:"ExpectedEndDate"` // Handle NULL
	IsOngoing       sql.NullBool   `json:"isOngoing,omitempty" db:"IsOngoing"`
}

// Skills defines the structure for the Skills table.
type Skills struct {
	Id       int64  `json:"ID" db:"Id"`
	PersonId int64  `json:"PersonId" db:"PersonId"`
	Skill    string `json:"Skill" db:"Skill"`
	Level    string `json:"Level" db:"Level"` // e.g., Basic, Intermediate, Advanced
}

// Languages defines the structure for the Languages table.
type Languages struct {
	Id       int64  `json:"ID" db:"Id"`
	PersonId int64  `json:"PersonId" db:"PersonId"`
	Language string `json:"Language" db:"Language"`
	Level    string `json:"Level" db:"Level"` // e.g., A1, A2, B1, B2, C1, C2, Native
}

// Project defines the structure for the Project table.
type Project struct {
	Id              int64          `json:"id" db:"Id"`
	PersonID        int64          `json:"person_id" db:"PersonID"`
	Title           string         `json:"title" db:"Title"`
	Role            sql.NullString `json:"role" db:"Role"`
	Description     sql.NullString `json:"description" db:"Description"`
	Company         sql.NullString `json:"company" db:"Company"`
	Document        sql.NullString `json:"document" db:"Document"`
	ProjectStatus   sql.NullString `json:"project_status" db:"ProjectStatus"`
	StartDate       sql.NullTime   `json:"start_date" db:"StartDate"`
	ExpectedEndDate sql.NullTime   `json:"expected_end_date" db:"ExpectedEndDate"`
	IsOngoing       sql.NullBool   `json:"isOngoing,omitempty" db:"IsOngoing"`
}

// Enterprise defines the structure for the Enterprise table.
type Enterprise struct {
	Id           int64          `json:"id" db:"Id"`
	RIF          string         `json:"rif" db:"RIF"`
	CompanyName  string         `json:"companyName" db:"CompanyName"`
	Password     string         `json:"-" db:"Password"`
	CategoryId   sql.NullInt64  `json:"categoryId" db:"CategoryId"` // Handle NULL
	CategoryName string         `json:"categoryName,omitempty" db:"CategoryName"`
	Description  sql.NullString `json:"description" db:"Description"` // Handle NULL
	Location     sql.NullString `json:"location" db:"Location"`       // Handle NULL
	WebSite      sql.NullString `json:"website" db:"WebSite"`         // Handle NULL (asumiendo que puede ser NULL)
	Email        sql.NullString `json:"email" db:"Email"`             // Handle NULL (asumiendo que puede ser NULL)
	Phone        sql.NullString `json:"phone" db:"Phone"`             // Handle NULL
	Picture      sql.NullString `json:"picture" db:"Picture"`         // Handle NULL
	CreateAt     time.Time      `json:"createAt" db:"CreateAt"`
	UpdateAt     time.Time      `json:"updateAt" db:"UpdateAt"`
}

// --- Helper Structs ---

// RegistrationStep1 defines the data for the first step of user registration.
type RegistrationStep1 struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	UserName  string `json:"userName"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Password  string `json:"password"`
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

// CompanyRegistrationRequest defines the data for company registration.
type CompanyRegistrationRequest struct {
	CompanyName string `json:"companyName"`
	RIF         string `json:"rif"`
	Sector      string `json:"sector"`
	ContactName string `json:"contactName"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Password    string `json:"password"`
	Location    string `json:"location"`
}

// LoginRequest defines the structure for login requests.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse defines the structure for login responses.
type LoginResponse struct {
	Token string  `json:"token"`
	User  UserDTO `json:"user"`
}

// UserDTO defines a clean user structure for API responses without sql.Null types
type UserDTO struct {
	Id                 int64  `json:"id"`
	FirstName          string `json:"first_name"`
	LastName           string `json:"last_name"`
	UserName           string `json:"user_name"`
	Email              string `json:"email"`
	Phone              string `json:"phone,omitempty"`
	Sex                string `json:"sex,omitempty"`
	DocId              string `json:"doc_id,omitempty"`
	NationalityId      int    `json:"nationality_id,omitempty"`
	Birthdate          string `json:"birthdate,omitempty"` // Format: YYYY-MM-DD
	Picture            string `json:"picture,omitempty"`
	DegreeId           int64  `json:"degree_id,omitempty"`
	UniversityId       int64  `json:"university_id,omitempty"`
	RoleId             int    `json:"role_id"`
	StatusAuthorizedId int    `json:"status_authorized_id"`
	Summary            string `json:"summary,omitempty"`
	Address            string `json:"address,omitempty"`
	Github             string `json:"github,omitempty"`
	Linkedin           string `json:"linkedin,omitempty"`
	CreatedAt          string `json:"created_at,omitempty"`
	UpdatedAt          string `json:"updated_at,omitempty"`
}

// ToUserDTO converts a User model to a clean UserDTO for API responses
func (u *User) ToUserDTO() UserDTO {
	dto := UserDTO{
		Id:                 u.Id,
		UserName:           u.UserName,
		Email:              u.Email,
		RoleId:             u.RoleId,
		StatusAuthorizedId: u.StatusAuthorizedId,
	}

	// Handle sql.Null* fields
	if u.FirstName.Valid {
		dto.FirstName = u.FirstName.String
	}
	if u.LastName.Valid {
		dto.LastName = u.LastName.String
	}
	if u.Phone.Valid {
		dto.Phone = u.Phone.String
	}
	if u.Sex.Valid {
		dto.Sex = u.Sex.String
	}
	if u.DocId.Valid {
		dto.DocId = u.DocId.String
	}
	if u.NationalityId.Valid {
		dto.NationalityId = int(u.NationalityId.Int32)
	}
	if u.Birthdate.Valid {
		dto.Birthdate = u.Birthdate.Time.Format("2006-01-02")
	}
	if u.Picture.Valid {
		dto.Picture = u.Picture.String
	}
	if u.DegreeId.Valid {
		dto.DegreeId = u.DegreeId.Int64
	}
	if u.UniversityId.Valid {
		dto.UniversityId = u.UniversityId.Int64
	}
	if u.Summary.Valid {
		dto.Summary = u.Summary.String
	}
	if u.Address.Valid {
		dto.Address = u.Address.String
	}
	if u.Github.Valid {
		dto.Github = u.Github.String
	}
	if u.Linkedin.Valid {
		dto.Linkedin = u.Linkedin.String
	}

	return dto
}

// WebSocketMessage defines the generic structure for incoming WebSocket messages.
type WebSocketMessage struct {
	Type    string      `json:"type"` // Corresponds to MessageType* constants
	Payload interface{} `json:"payload"`
}

// UserBaseInfo contiene un subconjunto de información del usuario, útil para identificación básica.
type UserBaseInfo struct {
	ID        int64  `json:"id"`
	FirstName string `json:"firstName,omitempty"` // Usar string directamente, se manejará NULL en Scan
	LastName  string `json:"lastName,omitempty"`  // Usar string directamente
	UserName  string `json:"userName"`
	Picture   string `json:"picture,omitempty"` // Usar string directamente
	RoleId    int    `json:"roleId"`
}

// GetUserBaseInfo recupera la información básica de un usuario por su ID.
func GetUserBaseInfo(db *sql.DB, userID int64) (*UserBaseInfo, error) {
	user := &UserBaseInfo{}
	query := `SELECT Id, FirstName, LastName, UserName, Picture, RoleId FROM User WHERE Id = ?`

	// Variables para escanear campos que pueden ser NULL
	var firstName, lastName, picture sql.NullString

	err := db.QueryRow(query, userID).Scan(
		&user.ID,
		&firstName,
		&lastName,
		&user.UserName, // Asumimos que UserName no es NULL por UNIQUE
		&picture,
		&user.RoleId, // Asumimos que RoleId no es NULL
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Devuelve un error específico si el usuario no se encuentra
			return nil, fmt.Errorf("user with ID %d not found", userID)
		}
		// Devuelve otros errores de la base de datos
		return nil, fmt.Errorf("error querying base user info for ID %d: %w", userID, err)
	}

	// Asignar valores desde sql.NullString a string (será "" si es NULL)
	user.FirstName = firstName.String
	user.LastName = lastName.String
	user.Picture = picture.String

	return user, nil
}

// GetDefaultNationalities devuelve una lista de nacionalidades por defecto

// ChatMessage representa un mensaje en la tabla ChatMessage de la base de datos.
type ChatMessage struct {
	ID         int64     `json:"id"`
	ChatID     string    `json:"chatId,omitempty"`   // Opcional, si tienes una tabla de Chat separada o para identificar el chat (ej. "user_1_user_2")
	FromUserID int64     `json:"fromUserId"`         // ID del remitente (tabla User)
	ToUserID   int64     `json:"toUserId"`           // ID del destinatario (tabla User)
	Content    string    `json:"content"`            // Contenido del mensaje
	CreatedAt  time.Time `json:"createdAt"`          // Timestamp de creación del mensaje (UTC)
	StatusID   int       `json:"statusId,omitempty"` // ID del estado del mensaje (ej. 1:sent, 2:delivered, 3:read) (tabla MessageStatus)
	// Otros campos como ClientTempID, ReplyToMessageID, etc. pueden ser añadidos.
}

// MessageStatus representa un estado de mensaje en la tabla MessageStatus.
type MessageStatus struct {
	ID   int    `json:"id"`
	Name string `json:"name"` // e.g., "sent", "delivered_to_server", "delivered_to_recipient_device", "read_by_recipient"
}

type UpdateProfilePayload struct {
	FirstName      *string `json:"firstName,omitempty"`
	LastName       *string `json:"lastName,omitempty"`
	UserName       *string `json:"userName,omitempty"`
	Phone          *string `json:"phone,omitempty"`
	Sex            *string `json:"sex,omitempty"`
	Birthdate      *string `json:"birthdate,omitempty"` // Formato esperado: YYYY-MM-DD
	NationalityID  *int64  `json:"nationalityId,omitempty"`
	Summary        *string `json:"summary,omitempty"`
	Address        *string `json:"address,omitempty"`
	Github         *string `json:"github,omitempty"`
	Linkedin       *string `json:"linkedin,omitempty"`
	CompanyName    *string `json:"companyName,omitempty"`
	Picture        *string `json:"picture,omitempty"`
	Email          *string `json:"email,omitempty"`
	ContactEmail   *string `json:"contactEmail,omitempty"`
	Twitter        *string `json:"twitter,omitempty"`
	Facebook       *string `json:"facebook,omitempty"`
	DocId          *string `json:"docId,omitempty"`
	DegreeId       *int64  `json:"degreeId,omitempty"`
	UniversityId   *int64  `json:"universityId,omitempty"`
	Sector         *string `json:"sector,omitempty"`
	Location       *string `json:"location,omitempty"`
	FoundationYear *int    `json:"foundationYear,omitempty"`
	EmployeeCount  *int    `json:"employeeCount,omitempty"`
}

// --- Profile View Structs ---

// UserProfile es la estructura para el perfil de usuario que se envía al frontend.
// Se diferencia del modelo User porque omite campos sensibles y maneja los tipos de datos
// para que coincidan con lo que espera el cliente (ej. `time.Time` en lugar de `sql.NullTime`).
type UserProfile struct {
	Id                 int64      `json:"id" db:"Id"`
	FirstName          string     `json:"firstName,omitempty" db:"FirstName"`
	LastName           string     `json:"lastName,omitempty" db:"LastName"`
	UserName           string     `json:"userName,omitempty" db:"UserName"`
	Email              string     `json:"email" db:"Email"`
	ContactEmail       string     `json:"contactEmail,omitempty" db:"ContactEmail"`
	Twitter            string     `json:"twitter,omitempty" db:"Twitter"`
	Facebook           string     `json:"facebook,omitempty" db:"Facebook"`
	Phone              string     `json:"phone,omitempty" db:"Phone"`
	Sex                string     `json:"sex,omitempty" db:"Sex"`
	DocId              string     `json:"docId,omitempty" db:"DocId"`
	NationalityId      *int32     `json:"nationalityId,omitempty" db:"NationalityId"`
	Birthdate          *time.Time `json:"birthdate,omitempty" db:"Birthdate"`
	Picture            string     `json:"picture,omitempty" db:"Picture"`
	DegreeId           *int64     `json:"degreeId,omitempty" db:"DegreeId"`
	UniversityId       *int64     `json:"universityId,omitempty" db:"UniversityId"`
	RoleId             int        `json:"roleId" db:"RoleId"`
	StatusAuthorizedId int        `json:"statusAuthorizedId,omitempty" db:"StatusAuthorizedId"`
	Summary            string     `json:"summary,omitempty" db:"Summary"`
	Address            string     `json:"address,omitempty" db:"Address"`
	Github             string     `json:"github,omitempty" db:"Github"`
	Linkedin           string     `json:"linkedin,omitempty" db:"Linkedin"`
	RIF                string     `json:"rif,omitempty" db:"RIF"`
	Sector             string     `json:"sector,omitempty" db:"Sector"`
	CompanyName        string     `json:"companyName,omitempty" db:"CompanyName"`
	Location           string     `json:"location,omitempty" db:"Location"`
	FoundationYear     *int       `json:"foundationYear,omitempty" db:"FoundationYear"`
	EmployeeCount      *int       `json:"employeeCount,omitempty" db:"EmployeeCount"`
	CreatedAt          time.Time  `json:"createdAt" db:"CreatedAt"`
	UpdatedAt          time.Time  `json:"updatedAt" db:"UpdatedAt"`
}

// CompleteProfile agrupa toda la información del perfil de un usuario.
type CompleteProfile struct {
	User           UserProfile      `json:"user"`
	Education      []Education      `json:"education"`
	WorkExperience []WorkExperience `json:"workExperience"`
	Certifications []Certifications `json:"certifications"`
	Skills         []Skills         `json:"skills"`
	Languages      []Languages      `json:"languages"`
	Projects       []Project        `json:"projects"`
	Degree         *Degree          `json:"degree,omitempty"`
}
