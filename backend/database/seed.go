package database

import (
	"time"

	"MendoCultura/models"
	"MendoCultura/services"

	"gorm.io/gorm"
)

func Seed(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		user, err := ensureUser(tx, "Usuario Demo", "usuario@mendocultura.local", "usuario123", "30111222", models.RoleUser)
		if err != nil {
			return err
		}
		_ = user

		organizer, err := ensureUser(tx, "Bodega Demo", "organizador@mendocultura.local", "organizador123", "30999888", models.RoleOrganizer)
		if err != nil {
			return err
		}
		if err := ensureOrganizerProfile(tx, organizer.ID); err != nil {
			return err
		}

		validator, err := ensureUser(tx, "Validador Demo", "validador@mendocultura.local", "validador123", "30777888", models.RoleValidator)
		if err != nil {
			return err
		}
		_ = validator

		admin, err := ensureUser(tx, "Admin Demo", "admin@mendocultura.local", "admin123", "30000000", models.RoleAdmin)
		if err != nil {
			return err
		}
		_ = admin

		if err := ensureEvents(tx, organizer.ID); err != nil {
			return err
		}

		return tx.FirstOrCreate(&models.AuditLog{}, models.AuditLog{Action: "SEED_CREATED", Entity: "system", Detail: "Datos de prueba iniciales creados"}).Error
	})
}

func ensureUser(tx *gorm.DB, name string, email string, password string, dni string, role string) (models.User, error) {
	var user models.User
	if err := tx.Where("email = ?", email).First(&user).Error; err == nil {
		return user, nil
	} else if err != gorm.ErrRecordNotFound {
		return models.User{}, err
	}

	hash, err := services.HashPassword(password)
	if err != nil {
		return models.User{}, err
	}

	user = models.User{
		Name:         name,
		Email:        email,
		PasswordHash: hash,
		DNI:          dni,
		Role:         role,
		Status:       models.AccountActive,
	}
	return user, tx.Create(&user).Error
}

func ensureOrganizerProfile(tx *gorm.DB, organizerID uint) error {
	var profile models.OrganizerProfile
	if err := tx.Where("user_id = ?", organizerID).First(&profile).Error; err == nil {
		return nil
	} else if err != gorm.ErrRecordNotFound {
		return err
	}

	profile = models.OrganizerProfile{
		UserID:       organizerID,
		BusinessName: "Bodega Demo Mendoza",
		TaxID:        "30-70000000-1",
		Locality:     "Luján de Cuyo",
		Status:       "APPROVED",
	}
	return tx.Create(&profile).Error
}

func ensureEvents(tx *gorm.DB, organizerID uint) error {
	events := []models.Event{
		{
			OrganizerID:         organizerID,
			Title:               "Fiesta de la Vendimia Joven",
			Description:         "Una noche joven con música cuyana, DJ sets, gastronomía regional y espíritu vendimial.",
			ExtendedDescription: "La Vendimia Joven reúne propuestas culturales, shows en vivo, patio gastronómico y espacios para disfrutar con amigos. Es una experiencia pensada para vivir Mendoza de manera fresca, segura y cercana.",
			Category:            "Vendimia",
			Department:          "Ciudad de Mendoza",
			Venue:               "Nave Cultural",
			Address:             "Juan Agustín Maza 250, Ciudad de Mendoza",
			ImageURL:            "https://images.unsplash.com/photo-1501281668745-f7f57925c3b4?auto=format&fit=crop&w=1200&q=80",
			GalleryImages:       "https://images.unsplash.com/photo-1492684223066-81342ee5ff30?auto=format&fit=crop&w=900&q=80,https://images.unsplash.com/photo-1506744038136-46273834b3fb?auto=format&fit=crop&w=900&q=80",
			TicketType:          "General",
			ImportantInfo:       "Ingreso por orden de llegada. Presentá DNI y QR de la entrada.",
			Recommendations:     "Llegá con 30 minutos de anticipación y llevá abrigo liviano.",
			StartAt:             time.Now().AddDate(0, 1, 2).Add(21 * time.Hour),
			PriceCents:          850000,
			Capacity:            220,
			AvailableTickets:    220,
			Status:              models.EventPublished,
		},
		{
			OrganizerID:         organizerID,
			Title:               "Noche de Bodegas en Luján",
			Description:         "Degustaciones, música acústica y cocina de estación entre viñedos de Luján de Cuyo.",
			ExtendedDescription: "Una experiencia turística y cultural para descubrir vinos mendocinos, recorrer espacios de bodega y disfrutar propuestas gastronómicas maridadas con Malbec.",
			Category:            "Bodegas",
			Department:          "Luján de Cuyo",
			Venue:               "Bodega Demo Mendoza",
			Address:             "Ruta Provincial 15, Luján de Cuyo, Mendoza",
			ImageURL:            "https://images.unsplash.com/photo-1506377247377-2a5b3b417ebb?auto=format&fit=crop&w=1200&q=80",
			GalleryImages:       "https://images.unsplash.com/photo-1510812431401-41d2bd2722f3?auto=format&fit=crop&w=900&q=80,https://images.unsplash.com/photo-1528823872057-9c018a7a7553?auto=format&fit=crop&w=900&q=80",
			TicketType:          "Degustación",
			ImportantInfo:       "Evento para mayores de 18 años. Incluye copa de bienvenida.",
			Recommendations:     "Usá calzado cómodo para recorrer sectores exteriores.",
			StartAt:             time.Now().AddDate(0, 1, 8).Add(20 * time.Hour),
			PriceCents:          1500000,
			Capacity:            80,
			AvailableTickets:    80,
			Status:              models.EventPublished,
		},
		{
			OrganizerID:         organizerID,
			Title:               "Festival de Música en Godoy Cruz",
			Description:         "Bandas mendocinas, food trucks y feria de emprendedores al aire libre.",
			ExtendedDescription: "El festival combina música local, espacios gastronómicos y propuestas para toda la familia en uno de los pulmones verdes de Godoy Cruz.",
			Category:            "Música",
			Department:          "Godoy Cruz",
			Venue:               "Espacio Verde Luis Menotti Pescarmona",
			Address:             "Balcarce y Rivadavia, Godoy Cruz, Mendoza",
			ImageURL:            "https://images.unsplash.com/photo-1470229722913-7c0e2dbbafd3?auto=format&fit=crop&w=1200&q=80",
			GalleryImages:       "https://images.unsplash.com/photo-1501386761578-eac5c94b800a?auto=format&fit=crop&w=900&q=80,https://images.unsplash.com/photo-1516450360452-9312f5e86fc7?auto=format&fit=crop&w=900&q=80",
			TicketType:          "Campo general",
			ImportantInfo:       "No se suspende por viento leve. Sector gastronómico con medios de pago digitales.",
			Recommendations:     "Llevá botella reutilizable y protector solar.",
			StartAt:             time.Now().AddDate(0, 2, 3).Add(18 * time.Hour),
			PriceCents:          700000,
			Capacity:            300,
			AvailableTickets:    300,
			Status:              models.EventPublished,
		},
		{
			OrganizerID:         organizerID,
			Title:               "Feria Gastronómica de Maipú",
			Description:         "Sabores regionales, productores locales, música y propuestas para toda la familia.",
			ExtendedDescription: "Un recorrido por la identidad gastronómica de Maipú con puestos de comida, degustaciones, cocina en vivo y mercado de productores mendocinos.",
			Category:            "Gastronomía",
			Department:          "Maipú",
			Venue:               "Parque Metropolitano Sur",
			Address:             "Ozamis Sur y Maza, Maipú, Mendoza",
			ImageURL:            "https://images.unsplash.com/photo-1555939594-58d7cb561ad1?auto=format&fit=crop&w=1200&q=80",
			GalleryImages:       "https://images.unsplash.com/photo-1504674900247-0877df9cc836?auto=format&fit=crop&w=900&q=80,https://images.unsplash.com/photo-1488459716781-31db52582fe9?auto=format&fit=crop&w=900&q=80",
			TicketType:          "Entrada general",
			ImportantInfo:       "Menores de 10 años ingresan sin cargo acompañados por un adulto.",
			Recommendations:     "Consultá el pronóstico y elegí ropa cómoda.",
			StartAt:             time.Now().AddDate(0, 0, 24).Add(17 * time.Hour),
			PriceCents:          450000,
			Capacity:            260,
			AvailableTickets:    260,
			Status:              models.EventPublished,
		},
		{
			OrganizerID:         organizerID,
			Title:               "Teatro de Verano en Capital",
			Description:         "Comedia mendocina y teatro independiente bajo las estrellas.",
			ExtendedDescription: "Una función especial para disfrutar producciones locales en formato de verano, con artistas invitados y charla breve posterior a la obra.",
			Category:            "Teatro",
			Department:          "Ciudad de Mendoza",
			Venue:               "Teatro Quintanilla",
			Address:             "Plaza Independencia, Ciudad de Mendoza",
			ImageURL:            "https://images.unsplash.com/photo-1503095396549-807759245b35?auto=format&fit=crop&w=1200&q=80",
			GalleryImages:       "https://images.unsplash.com/photo-1507924538820-ede94a04019d?auto=format&fit=crop&w=900&q=80,https://images.unsplash.com/photo-1460723237483-7a6dc9d0b212?auto=format&fit=crop&w=900&q=80",
			TicketType:          "Butaca numerada",
			ImportantInfo:       "La sala abre 40 minutos antes de la función.",
			Recommendations:     "Evitá llegar tarde: una vez iniciada la obra, el ingreso queda sujeto a disponibilidad.",
			StartAt:             time.Now().AddDate(0, 0, 20).Add(21 * time.Hour),
			PriceCents:          350000,
			Capacity:            120,
			AvailableTickets:    120,
			Status:              models.EventPublished,
		},
		{
			OrganizerID:         organizerID,
			Title:               "Experiencia Malbec en Tunuyán",
			Description:         "Atardecer en el Valle de Uco con degustación guiada y paisaje de montaña.",
			ExtendedDescription: "Una propuesta sensorial para conocer cepajes, historias de productores y vistas únicas del Valle de Uco en un entorno cuidado y relajado.",
			Category:            "Turismo",
			Department:          "Tunuyán",
			Venue:               "Finca Valle Demo",
			Address:             "Ruta Provincial 92, Tunuyán, Mendoza",
			ImageURL:            "https://images.unsplash.com/photo-1516594915697-87eb3b1c14ea?auto=format&fit=crop&w=1200&q=80",
			GalleryImages:       "https://images.unsplash.com/photo-1527661591475-527312dd65f5?auto=format&fit=crop&w=900&q=80,https://images.unsplash.com/photo-1500530855697-b586d89ba3ee?auto=format&fit=crop&w=900&q=80",
			TicketType:          "Experiencia guiada",
			ImportantInfo:       "Incluye degustación de tres etiquetas y tabla regional.",
			Recommendations:     "Ideal asistir con reserva de traslado si no contás con vehículo propio.",
			StartAt:             time.Now().AddDate(0, 1, 15).Add(19 * time.Hour),
			PriceCents:          2100000,
			Capacity:            60,
			AvailableTickets:    60,
			Status:              models.EventPublished,
		},
		{
			OrganizerID:         organizerID,
			Title:               "Congreso Universitario de Tecnología",
			Description:         "Charlas, workshops y networking para estudiantes y profesionales.",
			ExtendedDescription: "Una jornada para compartir proyectos, debatir tendencias y conectar con la comunidad tecnológica universitaria de Mendoza.",
			Category:            "Congresos",
			Department:          "Guaymallén",
			Venue:               "Centro Cultural Universitario",
			Address:             "Acceso Este, Guaymallén, Mendoza",
			ImageURL:            "https://images.unsplash.com/photo-1515187029135-18ee286d815b?auto=format&fit=crop&w=1200&q=80",
			GalleryImages:       "https://images.unsplash.com/photo-1540575467063-178a50c2df87?auto=format&fit=crop&w=900&q=80,https://images.unsplash.com/photo-1552664730-d307ca884978?auto=format&fit=crop&w=900&q=80",
			TicketType:          "Acreditación general",
			ImportantInfo:       "Incluye certificado digital de asistencia emitido por la organización.",
			Recommendations:     "Llevá notebook para los talleres prácticos.",
			StartAt:             time.Now().AddDate(0, 2, 12).Add(9 * time.Hour),
			PriceCents:          600000,
			Capacity:            180,
			AvailableTickets:    180,
			Status:              models.EventPublished,
		},
		{
			OrganizerID:         organizerID,
			Title:               "Partido Benéfico en San Rafael",
			Description:         "Encuentro deportivo solidario con artistas invitados y actividades familiares.",
			ExtendedDescription: "El partido reúne deporte, música y solidaridad. La recaudación simulada de la demo se muestra como ejemplo de trazabilidad para eventos con cupos controlados.",
			Category:            "Deportes",
			Department:          "San Rafael",
			Venue:               "Polideportivo Municipal",
			Address:             "Av. Hipólito Yrigoyen, San Rafael, Mendoza",
			ImageURL:            "https://images.unsplash.com/photo-1517927033932-b3d18e61fb3a?auto=format&fit=crop&w=1200&q=80",
			GalleryImages:       "https://images.unsplash.com/photo-1459865264687-595d652de67e?auto=format&fit=crop&w=900&q=80,https://images.unsplash.com/photo-1521412644187-c49fa049e84d?auto=format&fit=crop&w=900&q=80",
			TicketType:          "Popular",
			ImportantInfo:       "Ingreso con QR y DNI. Se recomienda llegar temprano por controles de acceso.",
			Recommendations:     "Traé alimento no perecedero si querés colaborar con la campaña solidaria.",
			StartAt:             time.Now().AddDate(0, 1, 26).Add(16 * time.Hour),
			PriceCents:          300000,
			Capacity:            500,
			AvailableTickets:    500,
			Status:              models.EventPublished,
		},
	}

	for _, event := range events {
		var existing models.Event
		if err := tx.Where("title = ?", event.Title).First(&existing).Error; err == nil {
			continue
		} else if err != gorm.ErrRecordNotFound {
			return err
		}
		if err := tx.Create(&event).Error; err != nil {
			return err
		}
	}
	return nil
}
