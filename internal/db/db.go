package db

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/davidM20/micro-service-backend-go.git/internal/models" // Ajusta la ruta si es necesario
	_ "github.com/go-sql-driver/mysql"
)

var (
	db   *sql.DB
	once sync.Once
)

// Connect initializes the database connection.
// It expects the DSN (Data Source Name) for the MySQL database.
func Connect(dsn string) (*sql.DB, error) {
	var err error
	once.Do(func() {
		db, err = sql.Open("mysql", dsn)
		if err != nil {
			return // err will be handled outside once.Do
		}

		db.SetConnMaxLifetime(time.Minute * 3)
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(10)

		err = db.Ping()
		if err != nil {
			db.Close()         // Close the connection if ping fails
			db = nil           // Reset db to nil so subsequent calls can retry
			once = sync.Once{} // Reset once so connection can be retried
			return
		}
		log.Println("Database connection successful!")
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	if db == nil { // Check if connection failed inside once.Do
		return nil, fmt.Errorf("database connection failed and was reset")
	}
	return db, nil
}

// GetDB returns the existing database connection pool.
// It's recommended to call Connect first.
func GetDB() *sql.DB {
	if db == nil {
		log.Println("Warning: GetDB called before Connect or connection failed.")
		// Depending on the strategy, you might want to panic or attempt reconnection here.
	}
	return db
}

// InitializeDatabase creates tables if they don't exist and populates default data.
func InitializeDatabase(conn *sql.DB) error {
	if conn == nil {
		return fmt.Errorf("database connection is nil")
	}

	tx, err := conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if anything fails

	if err := createTables(tx); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	if err := insertDefaultData(tx); err != nil {
		return fmt.Errorf("failed to insert default data: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Println("Database initialized successfully.")
	return nil
}

// createTables executes the CREATE TABLE IF NOT EXISTS statements.
func createTables(tx *sql.Tx) error {
	sqlSchema := `
    CREATE TABLE IF NOT EXISTS Token (
        Id INT PRIMARY KEY,
        TokenType VARCHAR(255) UNIQUE NOT NULL
    );

    CREATE TABLE IF NOT EXISTS Category (
        CategoryId BIGINT AUTO_INCREMENT PRIMARY KEY,
        Name VARCHAR(255),
        Description VARCHAR(255)
    );

    CREATE TABLE IF NOT EXISTS Interest (
        InterestId BIGINT AUTO_INCREMENT PRIMARY KEY,
        CategoryId BIGINT,
        Description VARCHAR(255),
        ExperienceLevel VARCHAR(255),
        FOREIGN KEY (CategoryId) REFERENCES Category(CategoryId)
    );

    CREATE TABLE IF NOT EXISTS TypeMessage (
        Id BIGINT AUTO_INCREMENT PRIMARY KEY,
        Name VARCHAR(255),
        Description VARCHAR(255)
    );

    CREATE TABLE IF NOT EXISTS Nationality (
        Id INT AUTO_INCREMENT PRIMARY KEY,
        CountryName VARCHAR(255) UNIQUE,
        IsoCode VARCHAR(255),
        DocIdFormat VARCHAR(255)
    );

    CREATE TABLE IF NOT EXISTS StatusAuthorized (
        Id INT PRIMARY KEY,
        Name VARCHAR(255) UNIQUE
    );

    CREATE TABLE IF NOT EXISTS University (
        Id BIGINT AUTO_INCREMENT PRIMARY KEY,
        Name VARCHAR(255) UNIQUE,
        Campus VARCHAR(255)
    );

    CREATE TABLE IF NOT EXISTS Degree (
        Id BIGINT AUTO_INCREMENT PRIMARY KEY,
        DegreeName VARCHAR(255),
        Descriptions VARCHAR(255),
        Code VARCHAR(255),
        UniversityId BIGINT,
        FOREIGN KEY (UniversityId) REFERENCES University(Id)
    );

    CREATE TABLE IF NOT EXISTS Role (
        Id INT PRIMARY KEY,
        Name VARCHAR(255) UNIQUE
    );

    CREATE TABLE IF NOT EXISTS User (
        Id BIGINT AUTO_INCREMENT PRIMARY KEY,
        FirstName VARCHAR(255),
        LastName VARCHAR(255),
        UserName VARCHAR(255) UNIQUE,
        Password VARCHAR(255),
        Email VARCHAR(255) UNIQUE NOT NULL,
        Phone VARCHAR(255),
        Sex VARCHAR(255),
        DocId VARCHAR(255) UNIQUE,
        NationalityId INT,
        Birthdate DATE,
        Picture VARCHAR(255),
        DegreeId BIGINT,
        UniversityId BIGINT,
        RoleId INT,
        StatusAuthorizedId INT,
        Summary VARCHAR(255),
        Address VARCHAR(255),
        Github VARCHAR(255),
        Linkedin VARCHAR(255),
        FOREIGN KEY (NationalityId) REFERENCES Nationality(Id),
        FOREIGN KEY (DegreeId) REFERENCES Degree(Id),
        FOREIGN KEY (UniversityId) REFERENCES University(Id),
        FOREIGN KEY (RoleId) REFERENCES Role(Id),
        FOREIGN KEY (StatusAuthorizedId) REFERENCES StatusAuthorized(Id)
    );

    CREATE TABLE IF NOT EXISTS Online (
        UserOnlineId BIGINT PRIMARY KEY,
        CreateAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Changed DATE to TIMESTAMP for more precision
        Status TINYINT(1),
        FOREIGN KEY (UserOnlineId) REFERENCES User(Id) ON DELETE CASCADE -- Added ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS Contact (
        ContactId BIGINT AUTO_INCREMENT PRIMARY KEY,
        User1Id BIGINT,
        User2Id BIGINT,
        Status VARCHAR(255),
        ChatId VARCHAR(255) UNIQUE,
        FOREIGN KEY (User1Id) REFERENCES User(Id) ON DELETE CASCADE, -- Added ON DELETE CASCADE
        FOREIGN KEY (User2Id) REFERENCES User(Id) ON DELETE CASCADE -- Added ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS Multimedia (
        Id VARCHAR(255) PRIMARY KEY, -- Use UUID generated by application
        Type VARCHAR(255),
        Ratio FLOAT,
        UserId BIGINT,
        FileName VARCHAR(255),
        CreateAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Changed DATE to TIMESTAMP
        ContentId VARCHAR(255), -- ID in cloud storage
        ChatId VARCHAR(255), -- Can be null if not associated with a chat
        FOREIGN KEY (UserId) REFERENCES User(Id) ON DELETE SET NULL -- Or CASCADE?
    );

    CREATE TABLE IF NOT EXISTS Session (
        Id BIGINT AUTO_INCREMENT PRIMARY KEY,
        UserId BIGINT,
        Tk TEXT, -- Changed VARCHAR(255) to TEXT for potentially longer JWTs
        Ip VARCHAR(255),
        RoleId INT,
        TokenId INT,
        CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        ExpiresAt TIMESTAMP, -- Added ExpiresAt for session management
        FOREIGN KEY (UserId) REFERENCES User(Id) ON DELETE CASCADE, -- Added ON DELETE CASCADE
        FOREIGN KEY (RoleId) REFERENCES Role(Id),
        FOREIGN KEY (TokenId) REFERENCES Token(Id)
    );

    CREATE TABLE IF NOT EXISTS Message (
        Id VARCHAR(255) PRIMARY KEY, -- Use UUID generated by application
        TypeMessageId BIGINT,
        Text TEXT, -- Changed VARCHAR(255) to TEXT for longer messages
        MediaId VARCHAR(255),
        Date TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Changed DATE to TIMESTAMP
        StatusMessage INT, -- 1: sent, 2: delivered, 3: read
        UserId BIGINT,
        ChatId VARCHAR(255),
        ResponseTo VARCHAR(255), -- Message ID it's replying to
        FOREIGN KEY (TypeMessageId) REFERENCES TypeMessage(Id),
        FOREIGN KEY (MediaId) REFERENCES Multimedia(Id) ON DELETE SET NULL, -- Keep message if media deleted?
        FOREIGN KEY (UserId) REFERENCES User(Id) ON DELETE SET NULL, -- Keep message if user deleted?
        FOREIGN KEY (ChatId) REFERENCES Contact(ChatId) ON DELETE CASCADE, -- Delete messages if chat deleted
        FOREIGN KEY (ResponseTo) REFERENCES Message(Id) ON DELETE SET NULL -- Keep message if replied-to deleted?
    );

    CREATE TABLE IF NOT EXISTS Education (
        Id BIGINT AUTO_INCREMENT PRIMARY KEY,
        PersonId BIGINT,
        Institution VARCHAR(255),
        Degree VARCHAR(255),
        Campus VARCHAR(255),
        GraduationDate DATE,
        CountryId BIGINT, -- Consider FOREIGN KEY to Nationality(Id)?
        FOREIGN KEY (PersonId) REFERENCES User(Id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS WorkExperience (
        Id BIGINT AUTO_INCREMENT PRIMARY KEY,
        PersonId BIGINT,
        Company VARCHAR(255),
        Position VARCHAR(255),
        StartDate DATE,
        EndDate DATE,
        Description TEXT, -- Changed VARCHAR(255) to TEXT
        CountryId BIGINT, -- Consider FOREIGN KEY to Nationality(Id)?
        FOREIGN KEY (PersonId) REFERENCES User(Id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS Certifications (
        Id BIGINT AUTO_INCREMENT PRIMARY KEY,
        PersonId BIGINT,
        Certification VARCHAR(255),
        Institution VARCHAR(255),
        DateObtained DATE,
        FOREIGN KEY (PersonId) REFERENCES User(Id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS Skills (
        Id BIGINT AUTO_INCREMENT PRIMARY KEY,
        PersonId BIGINT,
        Skill VARCHAR(255),
        Level VARCHAR(255),
        FOREIGN KEY (PersonId) REFERENCES User(Id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS Languages (
        Id BIGINT AUTO_INCREMENT PRIMARY KEY,
        PersonId BIGINT,
        Language VARCHAR(255),
        Level VARCHAR(255),
        FOREIGN KEY (PersonId) REFERENCES User(Id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS Project (
        Id BIGINT AUTO_INCREMENT PRIMARY KEY,
        PersonID BIGINT,
        Title VARCHAR(255),
        Role VARCHAR(255),
        Description TEXT, -- Changed VARCHAR(255) to TEXT
        Company VARCHAR(255),
        Document VARCHAR(255),
        ProjectStatus VARCHAR(255),
        StartDate DATE,
        ExpectedEndDate DATE,
        FOREIGN KEY (PersonID) REFERENCES User(Id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS Event (
        Id BIGINT AUTO_INCREMENT PRIMARY KEY,
        Description VARCHAR(255),
        UserId BIGINT, -- User receiving the notification
        OtherUserId BIGINT, -- User causing the event
        ProyectId BIGINT,
        CreateAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Changed DATE to TIMESTAMP
        IsRead TINYINT(1) DEFAULT 0, -- Added IsRead field
        FOREIGN KEY (UserId) REFERENCES User(Id) ON DELETE CASCADE,
        FOREIGN KEY (OtherUserId) REFERENCES User(Id) ON DELETE SET NULL, -- Keep event if other user deleted
        FOREIGN KEY (ProyectId) REFERENCES Project(Id) ON DELETE SET NULL -- Keep event if project deleted
    );

    CREATE TABLE IF NOT EXISTS Enterprise (
        Id BIGINT AUTO_INCREMENT PRIMARY KEY,
        RIF VARCHAR(255) UNIQUE NOT NULL,
        CompanyName VARCHAR(255) NOT NULL,
        CategoryId BIGINT,
        Description TEXT,
        Location VARCHAR(255),
        Phone VARCHAR(255),
        FOREIGN KEY (CategoryId) REFERENCES Category(CategoryId)
    );
	`

	_, err := tx.Exec(sqlSchema)
	if err != nil {
		return fmt.Errorf("error executing schema creation: %w", err)
	}
	log.Println("Tables created or already exist.")
	return nil
}

// insertDefaultData populates tables with initial values, ignoring duplicates.
func insertDefaultData(tx *sql.Tx) error {
	log.Println("Inserting default data...")

	// Insert Nationalities
	stmtNat, err := tx.Prepare("INSERT IGNORE INTO Nationality (CountryName, IsoCode, DocIdFormat) VALUES (?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare Nationality statement: %w", err)
	}
	defer stmtNat.Close()
	for _, nat := range models.GetDefaultNationalities() {
		_, err := stmtNat.Exec(nat.CountryName, nat.IsoCode, nat.DocIdFormat)
		if err != nil {
			log.Printf("Failed to insert nationality %s: %v", nat.CountryName, err)
			// Continue trying to insert others
		}
	}

	// Insert StatusAuthorized
	stmtStatus, err := tx.Prepare("INSERT IGNORE INTO StatusAuthorized (Id, Name) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare StatusAuthorized statement: %w", err)
	}
	defer stmtStatus.Close()
	for _, status := range models.GetDefaultStatusAuthorized() {
		_, err := stmtStatus.Exec(status.Id, status.Name)
		if err != nil {
			log.Printf("Failed to insert status %s: %v", status.Name, err)
		}
	}

	// Insert TokensType
	stmtToken, err := tx.Prepare("INSERT IGNORE INTO Token (Id, TokenType) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare Token statement: %w", err)
	}
	defer stmtToken.Close()
	for _, token := range models.GetDefaultTokensType() {
		_, err := stmtToken.Exec(token.Id, token.TokenType)
		if err != nil {
			log.Printf("Failed to insert token type %s: %v", token.TokenType, err)
		}
	}

	// Insert Roles
	stmtRole, err := tx.Prepare("INSERT IGNORE INTO Role (Id, Name) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare Role statement: %w", err)
	}
	defer stmtRole.Close()
	for _, role := range models.GetDefaultRoles() {
		_, err := stmtRole.Exec(role.Id, role.Name)
		if err != nil {
			log.Printf("Failed to insert role %s: %v", role.Name, err)
		}
	}

	// Insert Universities
	stmtUni, err := tx.Prepare("INSERT IGNORE INTO University (Name, Campus) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare University statement: %w", err)
	}
	defer stmtUni.Close()
	uniIDs := make(map[string]int64) // Store inserted university IDs
	for _, uni := range models.GetDefaultUniversities() {
		res, err := stmtUni.Exec(uni.Name, uni.Campus)
		if err != nil {
			log.Printf("Failed to insert university %s: %v", uni.Name, err)
			// Try to fetch existing ID if insertion failed due to duplicate
			row := tx.QueryRow("SELECT Id FROM University WHERE Name = ?", uni.Name)
			var id int64
			if scanErr := row.Scan(&id); scanErr == nil {
				uniIDs[uni.Name] = id
			} else {
				log.Printf("Failed to fetch existing ID for university %s: %v", uni.Name, scanErr)
			}
		} else {
			id, err := res.LastInsertId()
			if err == nil {
				uniIDs[uni.Name] = id
			} else {
				log.Printf("Failed to get last insert ID for university %s: %v", uni.Name, err)
			}
		}
	}

	// Insert Degrees (Requires University IDs)
	stmtDegree, err := tx.Prepare("INSERT IGNORE INTO Degree (DegreeName, Descriptions, Code, UniversityId) VALUES (?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare Degree statement: %w", err)
	}
	defer stmtDegree.Close()
	for _, deg := range models.GetDefaultDegrees() {
		// Find the corresponding University Name from the list to get the ID
		var uniName string
		for _, u := range models.GetDefaultUniversities() {
			if u.Id == deg.UniversityId { // Assuming original UniversityId in Degree refers to the implicit ID/index in the default list
				// THIS IS FRAGILE. It's better if Degree default data references University Name directly.
				// For now, we assume the default Degrees UniversityId = 1 maps to the first University in the list.
				if len(models.GetDefaultUniversities()) > 0 {
					uniName = models.GetDefaultUniversities()[0].Name
				}
				break
			}
		}

		uniID, ok := uniIDs[uniName]
		if !ok || uniName == "" {
			log.Printf("Skipping degree %s: Could not find University ID for assumed name '%s' (original ID: %d)", deg.DegreeName, uniName, deg.UniversityId)
			continue
		}

		_, err := stmtDegree.Exec(deg.DegreeName, deg.Descriptions, deg.Code, uniID)
		if err != nil {
			log.Printf("Failed to insert degree %s: %v", deg.DegreeName, err)
		}
	}

	// Insert TypeMessages
	stmtMsgType, err := tx.Prepare("INSERT IGNORE INTO TypeMessage (Id, Name, Description) VALUES (?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare TypeMessage statement: %w", err)
	}
	defer stmtMsgType.Close()
	for _, msgType := range models.GetDefaultTypeMessages() {
		_, err := stmtMsgType.Exec(msgType.Id, msgType.Name, msgType.Description)
		if err != nil {
			log.Printf("Failed to insert message type %s: %v", msgType.Name, err)
		}
	}

	log.Println("Finished inserting default data.")
	return nil
}
