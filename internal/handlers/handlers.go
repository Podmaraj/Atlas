package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"edgecore/internal/config"
	"edgecore/internal/models"
	"edgecore/pkg/jwt"
	"edgecore/pkg/redis"
)

type Handler struct {
	db          *gorm.DB
	redisClient *redis.Client
	cfg         *config.Config
}

func NewHandler(db *gorm.DB, rc *redis.Client, cfg *config.Config) *Handler {
	return &Handler{
		db:          db,
		redisClient: rc,
		cfg:         cfg,
	}
}

// 1. AUTHENTICATION HANDLERS
func (h *Handler) Login(c *fiber.Ctx) error {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	var user models.User
	if err := h.db.Where("username = ?", body.Username).First(&user).Error; err != nil {
		// Fallback for default admin user if DB is fresh
		if body.Username == h.cfg.ControlPlane.AdminUsername && body.Password == h.cfg.ControlPlane.AdminPassword {
			token, _ := jwt.GenerateToken(uuid.Nil, uuid.Nil, body.Username, "superadmin", h.cfg.ControlPlane.JWTSecret, h.cfg.ControlPlane.JWTExpiration)
			return c.JSON(fiber.Map{"token": token, "user": fiber.Map{"username": body.Username, "role": "superadmin"}})
		}
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.Password)); err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	token, err := jwt.GenerateToken(user.ID, user.OrganizationID, user.Username, user.Role, h.cfg.ControlPlane.JWTSecret, h.cfg.ControlPlane.JWTExpiration)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to sign token"})
	}

	return c.JSON(fiber.Map{
		"token": token,
		"user":  user,
	})
}

// 2. SERVICES HANDLERS
func (h *Handler) GetServices(c *fiber.Ctx) error {
	var services []models.Service
	if err := h.db.Preload("Routes").Preload("Instances").Find(&services).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(services)
}

func (h *Handler) CreateService(c *fiber.Ctx) error {
	var svc models.Service
	if err := c.BodyParser(&svc); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.db.Create(&svc).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	h.publishEvent(redis.EventServiceUpdated, svc.ID.String())
	return c.Status(210).JSON(svc)
}

func (h *Handler) DeleteService(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.db.Delete(&models.Service{}, "id = ?", id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	h.publishEvent(redis.EventServiceDeleted, id)
	return c.JSON(fiber.Map{"status": "deleted"})
}

// 3. ROUTES HANDLERS
func (h *Handler) GetRoutes(c *fiber.Ctx) error {
	var routes []models.Route
	if err := h.db.Preload("Plugins").Find(&routes).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(routes)
}

func (h *Handler) CreateRoute(c *fiber.Ctx) error {
	var rt models.Route
	if err := c.BodyParser(&rt); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.db.Create(&rt).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	h.publishEvent(redis.EventRouteUpdated, rt.ID.String())
	return c.Status(201).JSON(rt)
}

func (h *Handler) DeleteRoute(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.db.Delete(&models.Route{}, "id = ?", id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	h.publishEvent(redis.EventRouteDeleted, id)
	return c.JSON(fiber.Map{"status": "deleted"})
}

// 4. PLUGINS HANDLERS
func (h *Handler) GetPlugins(c *fiber.Ctx) error {
	var plugins []models.Plugin
	if err := h.db.Find(&plugins).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(plugins)
}

func (h *Handler) CreatePlugin(c *fiber.Ctx) error {
	var p models.Plugin
	if err := c.BodyParser(&p); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.db.Create(&p).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	h.publishEvent(redis.EventPluginUpdated, p.ID.String())
	return c.Status(201).JSON(p)
}

// 5. API KEYS HANDLERS
func (h *Handler) GetApiKeys(c *fiber.Ctx) error {
	var keys []models.ApiKey
	if err := h.db.Find(&keys).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(keys)
}

func (h *Handler) CreateApiKey(c *fiber.Ctx) error {
	var key models.ApiKey
	if err := c.BodyParser(&key); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	if err := h.db.Create(&key).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	h.publishEvent(redis.EventApiKeyUpdated, key.ID.String())
	return c.Status(201).JSON(key)
}

// 6. GATEWAY NODES & HEALTH HANDLERS
func (h *Handler) GetNodes(c *fiber.Ctx) error {
	var nodes []models.GatewayNode
	if err := h.db.Find(&nodes).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(nodes)
}

func (h *Handler) NodeHeartbeat(c *fiber.Ctx) error {
	var node models.GatewayNode
	if err := c.BodyParser(&node); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	node.LastHeartbeat = time.Now()
	if err := h.db.Where("node_id = ?", node.NodeID).Assign(node).FirstOrCreate(&node).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(node)
}

func (h *Handler) publishEvent(eventType redis.ConfigEventType, targetID string) {
	if h.redisClient != nil {
		_ = h.redisClient.PublishConfigEvent(context.Background(), redis.ConfigEvent{
			Type:      eventType,
			TargetID:  targetID,
			Timestamp: time.Now(),
		})
	}
}
