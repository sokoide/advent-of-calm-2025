package usecase

import (
	"fmt"

	"github.com/sokoide/advent-of-calm-2025/internal/domain"
)

const numGateways = 2

// EcommerceBuilder constructs the reference CALM architecture.
type EcommerceBuilder struct{}

// Build returns the populated e-commerce architecture model.
func (EcommerceBuilder) Build() *domain.Architecture {
	arch := domain.NewArchitecture(
		"ecommerce-platform-architecture",
		"E-Commerce Order Processing Platform",
		"A complete architecture for an e-commerce order processing system.",
	)

	setupGeneralInfo(arch)
	addGlobalControls(arch)

	nodes := defineNodes(arch)
	links := wireComponents(arch, nodes)
	defineFlows(arch, nodes, links)
	arch.DefineNode("api-gateway-1", domain.Service, "API Gateway Instance 1", "Primary API Gateway instance.")
	arch.DefineNode("api-gateway-2", domain.Service, "API Gateway Instance 2", "Secondary API Gateway instance for high availability.")

	return arch
}

var (
	metaTier1 = map[string]any{"tier": "tier-1", "business-criticality": "high"}
	metaTier2 = map[string]any{"tier": "tier-2", "business-criticality": "high"}

	metaOpsPlatform = map[string]any{"owner": "platform-team", "oncall-slack": "#oncall-platform"}
	metaOpsOrders   = map[string]any{"owner": "orders-team", "oncall-slack": "#oncall-orders"}
	metaOpsInv      = map[string]any{"owner": "inventory-team", "oncall-slack": "#oncall-inventory"}
	metaOpsPayments = map[string]any{"owner": "payments-team", "oncall-slack": "#oncall-payments"}

	metaManagedSvc = map[string]any{"deployment-type": "managed-service"}
	metaContainer  = map[string]any{"deployment-type": "container"}

	metaDBA = map[string]any{
		"tech-owner":  "DBA Team",
		"dba-contact": "dba-team@example.com",
	}
)

// --- Containers ---

type nodesContainer struct {
	Customer     *domain.Node
	Admin        *domain.Node
	LB           *domain.Node
	Gateways     []*domain.Node
	OrderSvc     *domain.Node
	InventorySvc *domain.Node
	PaymentSvc   *domain.Node
	OrderQueue   *domain.Node
	OrderPrimary *domain.Node
	OrderReplica *domain.Node
	InventoryDB  *domain.Node
	Broker       *domain.Node
	DBCluster    *domain.Node
	System       *domain.Node
}

type linksContainer struct {
	CustomerToLB *domain.Relationship
	AdminToLB    *domain.Relationship
	LBToGW       map[string]*domain.ConnectionBuilder
	GWToOrder    map[string]*domain.ConnectionBuilder
	GWToInv      map[string]*domain.ConnectionBuilder
	OrderToPriDB *domain.ConnectionBuilder
	OrderToRepDB *domain.ConnectionBuilder
	OrderToQueue *domain.ConnectionBuilder
	OrderToInv   *domain.ConnectionBuilder
	QueueToPay   *domain.ConnectionBuilder
	InvToDB      *domain.ConnectionBuilder
}

func setupGeneralInfo(a *domain.Architecture) {
	a.ADRs = []string{
		"docs/adr/0001-use-message-queue-for-async-processing.md",
		"docs/adr/0002-use-oauth2-for-api-authentication.md",
	}

	generalMeta := map[string]any{
		"owner":       "Architecture Team",
		"version":     "1.0.0",
		"created":     "2025-12-26",
		"description": "E-commerce order processing platform architecture demo.",
		"tags":        []string{"ecommerce", "microservices", "orders"},
		"monitoring": map[string]any{
			"grafana-dashboard": "https://grafana.example.com/d/ecommerce-overview",
			"kibana-logs":       "https://kibana.example.com/app/discover#/ecommerce-*",
			"pagerduty-service": "https://pagerduty.example.com/services/ECOMMERCE",
			"statuspage":        "https://status.example.com",
			"metrics-retention": "30 days",
			"log-retention":     "90 days",
		},
	}

	for k, v := range generalMeta {
		a.AddMeta(k, v)
	}
}

func addGlobalControls(a *domain.Architecture) {
	a.AddControl("security", "Data encryption and secure communication requirements",
		domain.NewRequirement("https://internal-policy.example.com/security/encryption-at-rest", domain.NewSecurityConfig("AES-256", "all-data-stores")),
		domain.NewRequirementURL("https://internal-policy.example.com/security/tls-1-3-minimum", "https://configs.example.com/security/tls-config.yaml"),
	).AddControl("performance", "System-wide performance and scalability requirements",
		domain.NewRequirement("https://internal-policy.example.com/performance/response-time-sla", domain.NewPerformanceConfig(200, 100)),
		domain.NewRequirementURL("https://internal-policy.example.com/performance/availability-target", "https://configs.example.com/infra/ha-config.yaml"),
	).AddControl("high-availability", "System-wide uptime and availability requirements",
		domain.NewRequirement("https://internal-policy.example.com/resilience/availability-sla", domain.NewAvailabilityConfig(99.9, 60)),
	)
}

func defineNodes(a *domain.Architecture) *nodesContainer {
	nc := &nodesContainer{}

	nc.Customer = a.DefineNode("customer", domain.Actor, "Customer", "A user who browses and purchases products.", domain.WithOwner("marketing-team", "CC-1000"))

	nc.Admin = a.DefineNode("admin", domain.Actor, "Admin", "A staff member who manages products and orders.", domain.WithOwner("ops-team", "CC-1000"))

	nc.System = a.DefineNode("ecommerce-system", domain.System, "AAA-Commerce Platform", "The overall e-commerce system containing microservices.", domain.WithOwner("platform-team", "CC-2000"))

	nc.LB = a.DefineNode("load-balancer", domain.Service, "Load Balancer", "High-availability entry point that distributes traffic to API Gateways.", domain.WithOwner("platform-team", "CC-2000"),
		domain.WithMeta(domain.Merge(metaTier1, metaOpsPlatform, metaManagedSvc, map[string]any{
			"tech-owner":      "Network Team",
			"ha-enabled":      true,
			"health-endpoint": "/status",
			"runbook":         "https://runbooks.example.com/load-balancer",
			"dashboard":       "https://grafana.example.com/d/lb-metrics",
			"log-query":       "service:load-balancer AND error",
			"alerts":          []string{"LB-HighLatency", "LB-TargetGroupUnhealthy"},
		})))
	nc.LB.Interface("lb-https", "HTTPS").SetName("Public HTTPS Interface").SetPort(443).SetHost("api.shop.example.com")
	nc.LB.Interface("lb-to-gateway", "HTTP").SetDesc("Outbound to API Gateways")

	gwPerf := []domain.Requirement{
		domain.NewRequirementURL("https://internal-policy.example.com/performance/rate-limiting", "https://configs.example.com/gateway/rate-limits.yaml"),
		domain.NewRequirement("https://internal-policy.example.com/performance/caching-policy", map[string]any{"default-ttl-seconds": 300, "cache-control": "private"}),
	}

	for i := 1; i <= numGateways; i++ {
		id := fmt.Sprintf("api-gateway-%d", i)
		desc := "Primary API Gateway instance."
		if i == 2 {
			desc = "Secondary API Gateway instance for high availability."
		}
		gw := a.DefineNode(id, domain.Service, fmt.Sprintf("API Gateway Instance %d", i), desc,
			domain.WithOwner("platform-team", "CC-2000"),
			domain.WithControl("performance", "API Gateway rate limiting and caching requirements", gwPerf...),
			domain.WithMeta(domain.Merge(metaTier1, metaOpsPlatform, metaContainer, map[string]any{
				"tech-owner":      "Edge Team",
				"health-endpoint": "/health",
				"runbook":         "https://runbooks.example.com/api-gateway",
				"dashboard":       "https://grafana.example.com/d/gateway-overview",
				"log-query":       fmt.Sprintf("service:api-gateway AND instance:%d", i),
				"alerts":          []string{"Gateway-5xx-Rate", "Gateway-HighLatency"},
				"repository":      "https://github.com/example/api-gateway",
			})))
		gw.Interface(fmt.Sprintf("gateway-%d-http", i), "HTTP").SetName("HTTP Interface").SetPort(80)
		gw.Interface(fmt.Sprintf("order-client-%d", i), "REST")
		gw.Interface(fmt.Sprintf("inventory-client-%d", i), "REST")
		gw.Interface(fmt.Sprintf("gateway-%d-health", i), "HTTP").SetName("Health Check").SetPath("/health")
		nc.Gateways = append(nc.Gateways, gw)
	}

	orderFailures := []map[string]any{
		{"check": "Check connection pool metrics in Grafana dashboard", "escalation": "If persists > 5min, page DBA team", "likely-cause": "Database connection pool exhausted", "remediation": "Scale up service replicas or increase pool size", "symptom": "HTTP 503 errors"},
		{"check": "Check payment-service health and circuit breaker status", "escalation": "Contact payments-team if circuit breaker not triggering", "likely-cause": "Payment service degradation", "remediation": "Circuit breaker should open automatically; check fallback queue", "symptom": "High latency (>2s p99)"},
		{"check": "Verify inventory-service cache TTL and database replication lag", "escalation": "Contact platform-team for cache issues", "likely-cause": "Inventory service returning stale data", "remediation": "Clear inventory cache; check replica sync status", "symptom": "Order validation failures"},
	}

	nc.OrderSvc = a.DefineNode("order-service", domain.Service, "Order Service", "Handles order creation and lifecycle management.", domain.WithOwner("orders-team", "CC-3000"),
		domain.WithMeta(domain.Merge(metaTier1, metaOpsOrders, metaContainer, map[string]any{
			"tech-owner":      "Order Team",
			"health-endpoint": "/actuator/health",
			"runbook":         "https://runbooks.example.com/order-service",
			"dashboard":       "https://grafana.example.com/d/order-service-metrics",
			"log-query":       "app:order-service AND level:ERROR",
			"alerts":          []string{"OrderCreationFailureRate", "OrderDBConectionExhausted"},
			"repository":      "https://github.com/example/order-service",
			"failure-modes":   orderFailures,
			"sla-tier":        "tier-1",
		})),
		domain.WithControl("circuit-breaker", "Fault tolerance for downstream service calls", domain.NewRequirement("https://internal-policy.example.com/resilience/circuit-breaker-policy", domain.NewCircuitBreakerConfig(50, 30, 10))),
	)
	nc.OrderSvc.Interface("order-api", "REST").SetName("Order API").SetPort(8080)
	nc.OrderSvc.Interface("order-db-write-client", "JDBC").SetName("Order DB Write Client").SetDesc("Writes to the primary database.")
	nc.OrderSvc.Interface("order-db-read-client", "JDBC").SetName("Order DB Read Client").SetDesc("Reads from the replica database.")
	nc.OrderSvc.Interface("payment-publisher", "AMQP").SetDesc("Publishes order messages for payment processing.")
	nc.OrderSvc.Interface("order-inventory-client", "REST").SetDesc("Outbound connection to check inventory")
	nc.OrderSvc.Interface("order-health", "HTTP").SetName("Health Check").SetPath("/health")

	nc.InventorySvc = a.DefineNode("inventory-service", domain.Service, "Inventory Service", "Manages product stock levels.", domain.WithOwner("inventory-team", "CC-4000"),
		domain.WithMeta(domain.Merge(metaTier2, metaOpsInv, metaContainer, map[string]any{
			"tech-owner":      "Warehouse Team",
			"health-endpoint": "/health",
			"runbook":         "https://runbooks.example.com/inventory-service",
			"dashboard":       "https://grafana.example.com/d/inventory-metrics",
			"log-query":       "app:inventory-service",
			"alerts":          []string{"InventoryCacheInconsistency", "StockUpdateFailure"},
			"repository":      "https://github.com/example/inventory-service",
			"failure-modes": []map[string]any{
				{"check": "Check DB lock metrics and slow query log", "escalation": "Contact DBA team for lock contention", "likely-cause": "Deadlock on stock updates", "remediation": "Review transaction isolation level or retry logic", "symptom": "Inventory sync failures"},
				{"check": "Verify Redis/Memcached availability and evictions", "escalation": "Contact platform-team for cache infrastructure", "likely-cause": "Cache invalidation failure", "remediation": "Flush cache for affected products", "symptom": "Stale stock levels"},
			},
		})))
	nc.InventorySvc.Interface("inventory-api", "REST").SetName("Inventory API").SetPort(8081)
	nc.InventorySvc.Interface("inventory-db-client", "JDBC")
	nc.InventorySvc.Interface("inventory-health", "HTTP").SetName("Health Check").SetPath("/health")

	nc.PaymentSvc = a.DefineNode("payment-service", domain.Service, "Payment Service", "Integrates with external payment providers.", domain.WithOwner("payments-team", "CC-5000"),
		domain.WithMeta(domain.Merge(metaTier1, metaOpsPayments, map[string]any{
			"deployment-type": "serverless",
			"tech-owner":      "Payment Team",
			"health-endpoint": "/health",
			"runbook":         "https://runbooks.example.com/payment-service",
			"dashboard":       "https://grafana.example.com/d/payment-metrics",
			"log-query":       "app:payment-service",
			"alerts":          []string{"PaymentGatewayTimeout", "PCIViolationAttempt"},
			"repository":      "https://github.com/example/payment-service",
			"failure-modes": []map[string]any{
				{"check": "Verify external gateway status page", "escalation": "Escalate to provider support", "likely-cause": "External payment gateway latency", "remediation": "Enable aggressive retry for idempotent calls", "symptom": "Payment processing timeouts"},
				{"check": "Review access logs for unusual patterns", "escalation": "Contact security-team", "likely-cause": "API Key leaked or compromised", "remediation": "Rotate API keys immediately", "symptom": "Unauthorized transaction spikes"},
			},
		})),
		domain.WithControl("compliance", "PCI-DSS compliance for payment processing", domain.NewRequirementURL("https://www.pcisecuritystandards.org/documents/PCI-DSS-v4.0", "https://configs.example.com/compliance/pci-dss-config.json")),
	)
	nc.PaymentSvc.Interface("payment-api", "REST").SetName("Payment Processing API").SetPort(8082)
	nc.PaymentSvc.Interface("payment-consumer", "AMQP").SetDesc("Consumes order messages for payment processing.")
	nc.PaymentSvc.Interface("payment-health", "HTTP").SetName("Health Check").SetPath("/health")

	nc.Broker = a.DefineNode("message-broker", domain.System, "Message Broker (RabbitMQ)", "Central messaging system for failure isolation and async processing.", domain.WithOwner("platform-team", "CC-2000"),
		domain.WithMeta(domain.Merge(metaOpsPlatform, metaManagedSvc, map[string]any{
			"tech-owner":      "Platform Team",
			"tier":            "tier-1",
			"health-endpoint": "/health",
			"runbook":         "https://runbooks.example.com/rabbitmq",
			"dashboard":       "https://grafana.example.com/d/rabbitmq-overview",
			"log-query":       "service:rabbitmq",
			"alerts":          []string{"RabbitMQQueueSizeHigh", "RabbitMQConsumerDown"},
			"adr":             "docs/adr/0001-use-message-queue-for-async-processing.md",
		})))
	nc.Broker.Interface("amqp-port", "AMQP").SetPort(5672)

	nc.OrderQueue = a.DefineNode("order-queue", domain.Queue, "Order Payment Queue", "Buffer for orders awaiting payment processing.", domain.WithOwner("orders-team", "CC-3000"))

	nc.DBCluster = a.DefineNode("order-database-cluster", domain.System, "Order Database Cluster", "High-availability database cluster for order data.", domain.WithOwner("dba-team", "CC-3000"),
		domain.WithControl("failover", "Disaster recovery and failover targets", domain.NewRequirement("https://internal-policy.example.com/resilience/disaster-recovery-targets", domain.NewFailoverConfig(15, 5, true))),
	)

	nc.OrderPrimary = a.DefineNode("order-database-primary", domain.Database, "Order Database (Primary)", "Main writable database for orders.", domain.WithOwner("dba-team", "CC-3000"), domain.WithMeta(domain.Merge(metaDBA, metaManagedSvc, map[string]any{"backup-schedule": "daily at 02:00 UTC", "restore-time": "60 minutes", "data-classification": "PII", "replication-mode": "async", "role": "primary"})))
	nc.OrderPrimary.Interface("order-sql-primary", "JDBC").SetPort(5432).SetDB("orders_v1").SetHost("orders-primary.example.com")

	nc.OrderReplica = a.DefineNode("order-database-replica", domain.Database, "Order Database (Replica)", "Read-only replica for scaling read operations.", domain.WithOwner("dba-team", "CC-3000"), domain.WithMeta(domain.Merge(metaDBA, metaManagedSvc, map[string]any{"backup-schedule": "daily at 02:00 UTC", "restore-time": "60 minutes", "data-classification": "PII", "replication-mode": "async", "role": "replica"})))
	nc.OrderReplica.Interface("order-sql-replica", "JDBC").SetPort(5432).SetDB("orders_v1").SetHost("orders-replica.example.com")

	nc.InventoryDB = a.DefineNode("inventory-db", domain.Database, "Inventory Database", "Stores stock levels.", domain.WithOwner("dba-team", "CC-4000"), domain.WithMeta(domain.Merge(metaDBA, metaManagedSvc, map[string]any{"backup-schedule": "weekly at Sunday 03:00 UTC", "restore-time": "30 minutes"})))
	nc.InventoryDB.Interface("inventory-sql", "JDBC").SetPort(5432).SetDB("inventory_v1").SetHost("inventory-db.example.com")

	// Dynamic dependencies
	gwIDs := []string{}
	for _, gw := range nc.Gateways {
		gwIDs = append(gwIDs, gw.UniqueID)
		gw.AddMeta("dependencies", []string{nc.OrderSvc.UniqueID, nc.InventorySvc.UniqueID})
	}
	nc.LB.AddMeta("dependencies", gwIDs)
	nc.OrderSvc.AddMeta("dependencies", []string{nc.DBCluster.UniqueID, nc.InventorySvc.UniqueID, nc.Broker.UniqueID})
	nc.InventorySvc.AddMeta("dependencies", []string{nc.InventoryDB.UniqueID})
	nc.PaymentSvc.AddMeta("dependencies", []string{"external-payment-provider"})

	return nc
}

func wireComponents(a *domain.Architecture, n *nodesContainer) *linksContainer {
	lc := &linksContainer{
		LBToGW:    make(map[string]*domain.ConnectionBuilder),
		GWToOrder: make(map[string]*domain.ConnectionBuilder),
		GWToInv:   make(map[string]*domain.ConnectionBuilder),
	}

	lc.CustomerToLB = a.Interacts("customer-interacts-lb", "Customer accesses the platform via Load Balancer.", n.Customer.UniqueID, n.LB.UniqueID).Data("public", true)
	lc.AdminToLB = a.Interacts("admin-interacts-lb", "Admin manages the platform via Load Balancer.", n.Admin.UniqueID, n.LB.UniqueID).Data("internal", true)

	for i, gw := range n.Gateways {
		id := gw.UniqueID
		lc.LBToGW[id] = n.LB.ConnectTo(gw, fmt.Sprintf("Load Balancer distributes traffic to Gateway Instance %d.", i+1)).
			WithID(fmt.Sprintf("lb-connects-gateway-%d", i+1)).Via("lb-to-gateway", fmt.Sprintf("gateway-%d-http", i+1)).Is("internal").Encrypted(true)
	}
	for i, gw := range n.Gateways {
		id := gw.UniqueID
		lc.GWToOrder[id] = gw.ConnectTo(n.OrderSvc, fmt.Sprintf("Gateway %d forwards requests to Order Service.", i+1)).
			WithID(fmt.Sprintf("gateway-%d-connects-order", i+1)).Via(fmt.Sprintf("order-client-%d", i+1), "order-api").Is("internal").Encrypted(true)
	}
	for i, gw := range n.Gateways {
		id := gw.UniqueID
		lc.GWToInv[id] = gw.ConnectTo(n.InventorySvc, fmt.Sprintf("Gateway %d forwards requests to Inventory Service.", i+1)).
			WithID(fmt.Sprintf("gateway-%d-connects-inventory", i+1)).Via(fmt.Sprintf("inventory-client-%d", i+1), "inventory-api").Is("internal").Encrypted(true)
	}

	lc.OrderToPriDB = n.OrderSvc.ConnectTo(n.OrderPrimary, "Order Service persists data to Primary Order Database.").
		WithID("order-connects-primary-db").Via("order-db-write-client", "order-sql-primary").Is("confidential").Encrypted(true).Tag("monitoring", true)
	lc.OrderToQueue = n.OrderSvc.ConnectTo(n.OrderQueue, "Order Service publishes payment task to queue.").
		WithID("order-publishes-to-queue").Via("payment-publisher", "").Is("internal").Encrypted(true).Protocol("AMQP")
	lc.QueueToPay = n.OrderQueue.ConnectTo(n.PaymentSvc, "Payment Service consumes payment task from queue.").
		WithID("payment-subscribes-to-queue").Via("", "payment-consumer").Is("internal").Encrypted(true).Protocol("AMQP")
	lc.OrderToRepDB = n.OrderSvc.ConnectTo(n.OrderReplica, "Order Service reads data from Replica Order Database.").
		WithID("order-connects-replica-db").Via("order-db-read-client", "order-sql-replica").Is("confidential").Encrypted(true).Tag("monitoring", true)
	lc.OrderToInv = n.OrderSvc.ConnectTo(n.InventorySvc, "Order Service checks/reserves stock in Inventory Service.").
		WithID("order-connects-inventory").Via("order-inventory-client", "inventory-api").Is("internal").Encrypted(true).
		Tag("monitoring", true).Tag("circuit-breaker", true).Tag("latency-sla", "< 100ms")
	lc.InvToDB = n.InventorySvc.ConnectTo(n.InventoryDB, "Inventory Service manages stock in Inventory Database.").
		WithID("inventory-connects-db").Via("inventory-db-client", "inventory-sql").Is("internal").Encrypted(true).Tag("monitoring", true)

	a.ComposedOf("broker-composition", "Message broker contains the order queue.", n.Broker.UniqueID, []string{n.OrderQueue.UniqueID}).Data("internal", false)
	a.ComposedOf("order-db-composition", "Primary and replica databases form the order database cluster.", n.DBCluster.UniqueID, []string{n.OrderPrimary.UniqueID, n.OrderReplica.UniqueID}).Data("confidential", true)

	sysNodes := []string{n.LB.UniqueID}
	for _, gw := range n.Gateways {
		sysNodes = append(sysNodes, gw.UniqueID)
	}
	sysNodes = append(sysNodes, n.OrderSvc.UniqueID, n.InventorySvc.UniqueID, n.PaymentSvc.UniqueID, n.DBCluster.UniqueID, n.InventoryDB.UniqueID, n.Broker.UniqueID)
	a.ComposedOf("ecommerce-system-composition", "The E-Commerce Platform comprises its core services and databases.", n.System.UniqueID, sysNodes).Data("internal", false)

	return lc
}

func defineFlows(a *domain.Architecture, n *nodesContainer, l *linksContainer) {
	if len(n.Gateways) == 0 {
		return
	}

	// List of flow "recipes" to be added
	flowTasks := []func(gw *domain.Node){
		func(gw *domain.Node) { addOrderFlow(a, gw, l) },
		func(gw *domain.Node) { addInventoryFlow(a, gw, l) },
	}

	// Distribute flows across available gateways using round-robin
	for i, task := range flowTasks {
		targetGW := n.Gateways[i%len(n.Gateways)]
		task(targetGW)
	}
}

func addOrderFlow(a *domain.Architecture, gw *domain.Node, l *linksContainer) {
	a.DefineFlow("order-processing-flow", "Customer Order Processing", "End-to-end flow from customer placing an order to payment confirmation").
		MetaMap(map[string]any{
			"business-impact":        "Customers cannot complete purchases - direct revenue loss",
			"degraded-behavior":      "Orders queue in message broker; processed when service recovers",
			"customer-communication": "Display 'Order processing delayed' message",
			"sla":                    "99.9% availability, 30s p99 latency",
		}).
		Steps(
			domain.StepSpec{ID: l.CustomerToLB.UniqueID, Desc: "Customer submits order via Load Balancer"},
			domain.StepSpec{ID: l.LBToGW[gw.UniqueID].GetID(), Desc: fmt.Sprintf("LB routes to %s", gw.Name)},
			domain.StepSpec{ID: l.GWToOrder[gw.UniqueID].GetID(), Desc: fmt.Sprintf("%s routes to Order Service", gw.Name)},
			domain.StepSpec{ID: l.OrderToQueue.GetID(), Desc: "Order Service publishes payment task"},
			domain.StepSpec{ID: l.QueueToPay.GetID(), Desc: "Payment Service processes task from queue"},
		)
}

func addInventoryFlow(a *domain.Architecture, gw *domain.Node, l *linksContainer) {
	a.DefineFlow("inventory-check-flow", "Inventory Stock Check", "Admin checks and updates inventory stock levels").
		MetaMap(map[string]any{
			"business-impact":        "Stock levels may be inaccurate - risk of overselling",
			"customer-communication": "Display 'Stock availability being confirmed'",
			"degraded-behavior":      "Fall back to cached inventory; flag orders for manual review",
			"sla":                    "99.5% availability, 500ms p99 latency",
		}).
		Steps(
			domain.StepSpec{ID: l.AdminToLB.UniqueID, Desc: "Admin requests inventory status via LB"},
			domain.StepSpec{ID: l.LBToGW[gw.UniqueID].GetID(), Desc: fmt.Sprintf("LB routes to %s", gw.Name)},
			domain.StepSpec{ID: l.GWToInv[gw.UniqueID].GetID(), Desc: fmt.Sprintf("%s routes to inventory service", gw.Name)},
			domain.StepSpec{ID: l.InvToDB.GetID(), Desc: "Query current stock levels"},
			domain.StepSpec{ID: l.InvToDB.GetID(), Desc: "Return stock data", Dir: "destination-to-source"},
			domain.StepSpec{ID: l.GWToInv[gw.UniqueID].GetID(), Desc: "Return inventory report", Dir: "destination-to-source"},
		)
}
