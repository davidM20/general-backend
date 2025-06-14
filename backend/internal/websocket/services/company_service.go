package services

import (
	"context"
	"time"

	"github.com/davidM20/micro-service-backend-go.git/internal/db/queries"
	"github.com/davidM20/micro-service-backend-go.git/internal/models"
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
	"golang.org/x/sync/errgroup"
)

// GetCompleteCompanyProfile reúne toda la información del perfil de una empresa.
func GetCompleteCompanyProfile(userID int64) (*models.CompleteCompanyProfile, error) {
	var completeProfile models.CompleteCompanyProfile
	g, _ := errgroup.WithContext(context.Background())

	// 1. Obtener el perfil de la empresa
	g.Go(func() error {
		companyProfile, err := queries.GetCompanyProfile(userID)
		if err != nil {
			logger.Errorf("COMPANY_SERVICE", "Error obteniendo el perfil para CompanyID %d: %v", userID, err)
			return err
		}
		completeProfile.Company = *companyProfile
		return nil
	})

	// 2. Obtener los eventos de la empresa
	g.Go(func() error {
		events, err := queries.GetEventsForCompany(userID)
		if err != nil {
			logger.Warnf("COMPANY_SERVICE", "Error obteniendo eventos para CompanyID %d: %v", userID, err)
			return nil // No es un error fatal, se puede mostrar el perfil sin eventos
		}
		completeProfile.Events = events
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	// 3. Calcular las estadísticas
	calculateCompanyStats(&completeProfile)

	return &completeProfile, nil
}

// calculateCompanyStats calcula y asigna las estadísticas de la empresa.
func calculateCompanyStats(profile *models.CompleteCompanyProfile) {
	totalEvents := len(profile.Events)
	var upcomingEvents int
	for _, event := range profile.Events {
		if event.EventDate.After(time.Now()) {
			upcomingEvents++
		}
	}

	var yearsOfExperience int
	if profile.Company.FoundationYear != nil && *profile.Company.FoundationYear > 0 {
		yearsOfExperience = time.Now().Year() - *profile.Company.FoundationYear
	}

	var employeeCount int
	if profile.Company.EmployeeCount != nil {
		employeeCount = *profile.Company.EmployeeCount
	}

	profile.Stats = models.CompanyStats{
		TotalEvents:       totalEvents,
		UpcomingEvents:    upcomingEvents,
		YearsOfExperience: yearsOfExperience,
		EmployeeCount:     employeeCount,
	}
}
