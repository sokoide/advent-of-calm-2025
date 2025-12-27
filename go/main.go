package main

import "fmt"

func main() {
	arch := NewArchitecture(
		"ecommerce-platform-architecture",
		"E-Commerce Order Processing Platform",
		"A complete architecture for an e-commerce order processing system.",
	)

	setupGeneralInfo(arch)
	addGlobalControls(arch)

	nodes := defineNodes(arch)
	wireComponents(arch, nodes)

	fmt.Println(arch.ToJSON())
}

type NodesContainer struct {
	Customer     *Node
	Admin        *Node
	LB           *Node
	Gateways     []*Node
	OrderSvc     *Node
	InventorySvc *Node
	PaymentSvc   *Node
	OrderQueue   *Node
	OrderPrimary *Node
	OrderReplica *Node
	InventoryDB  *Node
}

func setupGeneralInfo(a *Architecture) {
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

func addGlobalControls(a *Architecture) {
	a.AddControl(
		"security",
		"Data encryption and secure communication requirements",
		NewRequirement(
			"https://internal-policy.example.com/security/encryption-at-rest",
			NewSecurityConfig("AES-256", "all-data-stores"),
		),
		NewRequirementURL(
			"https://internal-policy.example.com/security/tls-1-3-minimum",
			"https://configs.example.com/security/tls-config.yaml",
		),
	).
		AddControl(
			"performance",
			"System-wide performance and scalability requirements",
			NewRequirement(
				"https://internal-policy.example.com/performance/response-time-sla",
				NewPerformanceConfig(200, 100),
			),
			NewRequirementURL(
				"https://internal-policy.example.com/performance/availability-target",
				"https://configs.example.com/infra/ha-config.yaml",
			),
		).
		AddControl(
			"high-availability",
			"System-wide uptime and availability requirements",
			NewRequirement(
				"https://internal-policy.example.com/resilience/availability-sla",
				NewAvailabilityConfig(99.9, 60),
			),
		)
}

func defineNodes(a *Architecture) *NodesContainer {
	nc := &NodesContainer{}

	nc.Customer = a.DefineNode("customer", Actor, "Customer", "A user who browses and purchases products.",
		WithOwner("marketing-team", "CC-1000"))

	nc.Admin = a.DefineNode("admin", Actor, "Admin", "A staff member who manages products and orders.",
		WithOwner("ops-team", "CC-1000"))

	a.DefineNode(
		"ecommerce-system",
		System,
		"E-Commerce Platform",
		"The overall e-commerce system containing microservices.",
		WithOwner("platform-team", "CC-2000"),
	)

	nc.LB = a.DefineNode(
		"load-balancer",
		Service,
		"Load Balancer",
		"High-availability entry point that distributes traffic to API Gateways.",
		WithOwner("platform-team", "CC-2000"),
		WithMeta(map[string]any{
			"tech-owner":           "Network Team",
			"owner":                "platform-team",
			"oncall-slack":         "#oncall-platform",
			"tier":                 "tier-1",
			"business-criticality": "high",
			"ha-enabled":           true,
			"health-endpoint":      "/status",
			"runbook":              "https://runbooks.example.com/load-balancer",
			"dashboard":            "https://grafana.example.com/d/lb-metrics",
			"log-query":            "service:load-balancer AND error",
			"alerts":               []string{"LB-HighLatency", "LB-TargetGroupUnhealthy"},
			"deployment-type":      "managed-service",
			"dependencies":         []string{"api-gateway-1", "api-gateway-2"},
		}),
	)
	nc.LB.Interface("lb-https", "HTTPS").SetName("Public HTTPS Interface").SetPort(443).SetHost("api.shop.example.com")
	nc.LB.Interface("lb-to-gateway", "HTTP").SetDesc("Outbound to API Gateways")

	gwPerf := []Requirement{
		NewRequirementURL(
			"https://internal-policy.example.com/performance/rate-limiting",
			"https://configs.example.com/gateway/rate-limits.yaml",
		),
		NewRequirement(
			"https://internal-policy.example.com/performance/caching-policy",
			map[string]any{"default-ttl-seconds": 300, "cache-control": "private"},
		),
	}

	for i := 1; i <= 2; i++ {
		id := fmt.Sprintf("api-gateway-%d", i)
		name := fmt.Sprintf("API Gateway Instance %d", i)
		desc := "Primary API Gateway instance."
		if i == 2 {
			desc = "Secondary API Gateway instance for high availability."
		}

		gw := a.DefineNode(id, Service, name, desc,
			WithOwner("platform-team", "CC-2000"),
			WithControl("performance", "API Gateway rate limiting and caching requirements", gwPerf...),
			WithMeta(map[string]any{
				"tech-owner":           "Edge Team",
				"owner":                "platform-team",
				"deployment-type":      "container",
				"health-endpoint":      "/health",
				"runbook":              "https://runbooks.example.com/api-gateway",
				"dashboard":            "https://grafana.example.com/d/gateway-overview",
				"log-query":            fmt.Sprintf("service:api-gateway AND instance:%d", i),
				"alerts":               []string{"Gateway-5xx-Rate", "Gateway-HighLatency"},
				"repository":           "https://github.com/example/api-gateway",
				"business-criticality": "high",
				"tier":                 "tier-1",
				"oncall-slack":         "#oncall-platform",
				"dependencies":         []string{"order-service", "inventory-service"},
			}))
		gw.Interface(fmt.Sprintf("gateway-%d-http", i), "HTTP").SetName("HTTP Interface").SetPort(80)
		gw.Interface(fmt.Sprintf("order-client-%d", i), "REST")
		gw.Interface(fmt.Sprintf("inventory-client-%d", i), "REST")
		gw.Interface(fmt.Sprintf("gateway-%d-health", i), "HTTP").SetName("Health Check").SetPath("/health")

		nc.Gateways = append(nc.Gateways, gw)
	}

	orderFailures := []map[string]any{
		{
			"check":        "Check connection pool metrics in Grafana dashboard",
			"escalation":   "If persists > 5min, page DBA team",
			"likely-cause": "Database connection pool exhausted",
			"remediation":  "Scale up service replicas or increase pool size",
			"symptom":      "HTTP 503 errors",
		},
		{
			"check":        "Check payment-service health and circuit breaker status",
			"escalation":   "Contact payments-team if circuit breaker not triggering",
			"likely-cause": "Payment service degradation",
			"remediation":  "Circuit breaker should open automatically; check fallback queue",
			"symptom":      "High latency (>2s p99)",
		},
		{
			"check":        "Verify inventory-service cache TTL and database replication lag",
			"escalation":   "Contact platform-team for cache issues",
			"likely-cause": "Inventory service returning stale data",
			"remediation":  "Clear inventory cache; check replica sync status",
			"symptom":      "Order validation failures",
		},
	}

	nc.OrderSvc = a.DefineNode(
		"order-service",
		Service,
		"Order Service",
		"Handles order creation and lifecycle management.",
		WithOwner("orders-team", "CC-3000"),
		WithMeta(map[string]any{
			"tier":                 "tier-1",
			"owner":                "orders-team",
			"business-criticality": "high",
			"tech-owner":           "Order Team",
			"oncall-slack":         "#oncall-orders",
			"health-endpoint":      "/actuator/health",
			"runbook":              "https://runbooks.example.com/order-service",
			"dashboard":            "https://grafana.example.com/d/order-service-metrics",
			"log-query":            "app:order-service AND level:ERROR",
			"alerts":               []string{"OrderCreationFailureRate", "OrderDBConectionExhausted"},
			"deployment-type":      "container",
			"repository":           "https://github.com/example/order-service",
			"failure-modes":        orderFailures,
			"dependencies":         []string{"order-database-cluster", "inventory-service", "message-broker"},
			"sla-tier":             "tier-1",
		}),
		WithControl(
			"circuit-breaker",
			"Fault tolerance for downstream service calls",
			NewRequirement(
				"https://internal-policy.example.com/resilience/circuit-breaker-policy",
				NewCircuitBreakerConfig(50, 30, 10),
			),
		),
	)
	nc.OrderSvc.Interface("order-api", "REST").SetName("Order API").SetPort(8080)
	nc.OrderSvc.Interface("order-db-write-client", "JDBC").
		SetName("Order DB Write Client").
		SetDesc("Writes to the primary database.")
	nc.OrderSvc.Interface("order-db-read-client", "JDBC").
		SetName("Order DB Read Client").
		SetDesc("Reads from the replica database.")
	nc.OrderSvc.Interface("payment-publisher", "AMQP").SetDesc("Publishes order messages for payment processing.")
	nc.OrderSvc.Interface("order-inventory-client", "REST").SetDesc("Outbound connection to check inventory")
	nc.OrderSvc.Interface("order-health", "HTTP").SetName("Health Check").SetPath("/health")

	nc.InventorySvc = a.DefineNode("inventory-service", Service, "Inventory Service", "Manages product stock levels.",
		WithOwner("inventory-team", "CC-4000"),
		WithMeta(map[string]any{
			"tier":                 "tier-2",
			"owner":                "inventory-team",
			"tech-owner":           "Warehouse Team",
			"oncall-slack":         "#oncall-inventory",
			"health-endpoint":      "/health",
			"runbook":              "https://runbooks.example.com/inventory-service",
			"dashboard":            "https://grafana.example.com/d/inventory-metrics",
			"log-query":            "app:inventory-service",
			"alerts":               []string{"InventoryCacheInconsistency", "StockUpdateFailure"},
			"deployment-type":      "container",
			"repository":           "https://github.com/example/inventory-service",
			"business-criticality": "high",
			"dependencies":         []string{"inventory-db"},
			"failure-modes": []map[string]any{
				{
					"check":        "Check DB lock metrics and slow query log",
					"escalation":   "Contact DBA team for lock contention",
					"likely-cause": "Deadlock on stock updates",
					"remediation":  "Review transaction isolation level or retry logic",
					"symptom":      "Inventory sync failures",
				},
				{
					"check":        "Verify Redis/Memcached availability and evictions",
					"escalation":   "Contact platform-team for cache infrastructure",
					"likely-cause": "Cache invalidation failure",
					"remediation":  "Flush cache for affected products",
					"symptom":      "Stale stock levels",
				},
			},
		}))
	nc.InventorySvc.Interface("inventory-api", "REST").SetName("Inventory API").SetPort(8081)
	nc.InventorySvc.Interface("inventory-db-client", "JDBC")
	nc.InventorySvc.Interface("inventory-health", "HTTP").SetName("Health Check").SetPath("/health")

	nc.PaymentSvc = a.DefineNode(
		"payment-service",
		Service,
		"Payment Service",
		"Integrates with external payment providers.",
		WithOwner("payments-team", "CC-5000"),
		WithMeta(map[string]any{
			"tier":                 "tier-1",
			"owner":                "payments-team",
			"deployment-type":      "serverless",
			"tech-owner":           "Payment Team",
			"oncall-slack":         "#oncall-payments",
			"health-endpoint":      "/health",
			"runbook":              "https://runbooks.example.com/payment-service",
			"dashboard":            "https://grafana.example.com/d/payment-metrics",
			"log-query":            "app:payment-service",
			"alerts":               []string{"PaymentGatewayTimeout", "PCIViolationAttempt"},
			"repository":           "https://github.com/example/payment-service",
			"business-criticality": "high",
			"dependencies":         []string{"external-payment-provider"},
			"failure-modes": []map[string]any{
				{
					"check":        "Verify external gateway status page",
					"escalation":   "Escalate to provider support",
					"likely-cause": "External payment gateway latency",
					"remediation":  "Enable aggressive retry for idempotent calls",
					"symptom":      "Payment processing timeouts",
				},
				{
					"check":        "Review access logs for unusual patterns",
					"escalation":   "Contact security-team",
					"likely-cause": "API Key leaked or compromised",
					"remediation":  "Rotate API keys immediately",
					"symptom":      "Unauthorized transaction spikes",
				},
			},
		}),
		WithControl(
			"compliance",
			"PCI-DSS compliance for payment processing",
			NewRequirementURL(
				"https://www.pcisecuritystandards.org/documents/PCI-DSS-v4.0",
				"https://configs.example.com/compliance/pci-dss-config.json",
			),
		),
	)
	nc.PaymentSvc.Interface("payment-api", "REST").SetName("Payment Processing API").SetPort(8082)
	nc.PaymentSvc.Interface("payment-consumer", "AMQP").SetDesc("Consumes order messages for payment processing.")
	nc.PaymentSvc.Interface("payment-health", "HTTP").SetName("Health Check").SetPath("/health")

	a.DefineNode("message-broker", System, "Message Broker (RabbitMQ)", "Central messaging system for failure isolation and async processing.",
		WithOwner("platform-team", "CC-2000"),
		WithMeta(map[string]any{
			"tech-owner":      "Platform Team",
			"owner":           "platform-team",
			"tier":            "tier-1",
			"oncall-slack":    "#oncall-platform",
			"health-endpoint": "/health",
			"runbook":         "https://runbooks.example.com/rabbitmq",
			"dashboard":       "https://grafana.example.com/d/rabbitmq-overview",
			"log-query":       "service:rabbitmq",
			"alerts":          []string{"RabbitMQQueueSizeHigh", "RabbitMQConsumerDown"},
			"deployment-type": "managed-service",
			"adr":             "docs/adr/0001-use-message-queue-for-async-processing.md",
		})).
		Interface("amqp-port", "AMQP").
		SetPort(5672)

	nc.OrderQueue = a.DefineNode(
		"order-queue",
		Queue,
		"Order Payment Queue",
		"Buffer for orders awaiting payment processing.",
		WithOwner("orders-team", "CC-3000"),
	)

	a.DefineNode(
		"order-database-cluster",
		System,
		"Order Database Cluster",
		"High-availability database cluster for order data.",
		WithOwner("dba-team", "CC-3000"),
		WithControl(
			"failover",
			"Disaster recovery and failover targets",
			NewRequirement(
				"https://internal-policy.example.com/resilience/disaster-recovery-targets",
				NewFailoverConfig(15, 5, true),
			),
		),
	)

	dbCommonMeta := map[string]any{
		"tech-owner":          "DBA Team",
		"backup-schedule":     "daily at 02:00 UTC",
		"restore-time":        "60 minutes",
		"dba-contact":         "dba-team@example.com",
		"deployment-type":     "managed-service",
		"data-classification": "PII",
		"replication-mode":    "async",
	}

	nc.OrderPrimary = a.DefineNode(
		"order-database-primary",
		Database,
		"Order Database (Primary)",
		"Main writable database for orders.",
		WithOwner("dba-team", "CC-3000"),
		WithMeta(dbCommonMeta),
		WithMeta(map[string]any{"role": "primary"}),
	)
	nc.OrderPrimary.Interface("order-sql-primary", "JDBC").
		SetPort(5432).
		SetDB("orders_v1").
		SetHost("orders-primary.example.com")

	nc.OrderReplica = a.DefineNode(
		"order-database-replica",
		Database,
		"Order Database (Replica)",
		"Read-only replica for scaling read operations.",
		WithOwner("dba-team", "CC-3000"),
		WithMeta(dbCommonMeta),
		WithMeta(map[string]any{"role": "replica"}),
	)
	nc.OrderReplica.Interface("order-sql-replica", "JDBC").
		SetPort(5432).
		SetDB("orders_v1").
		SetHost("orders-replica.example.com")

	nc.InventoryDB = a.DefineNode("inventory-db", Database, "Inventory Database", "Stores stock levels.",
		WithOwner("dba-team", "CC-4000"), WithMeta(map[string]any{
			"tech-owner":      "DBA Team",
			"backup-schedule": "weekly at Sunday 03:00 UTC",
			"restore-time":    "30 minutes",
			"dba-contact":     "dba-team@example.com",
			"deployment-type": "managed-service",
		}))
	nc.InventoryDB.Interface("inventory-sql", "JDBC").
		SetPort(5432).
		SetDB("inventory_v1").
		SetHost("inventory-db.example.com")

	return nc
}

func wireComponents(a *Architecture, n *NodesContainer) {
	// --- Interactions ---
	customerInteractsLB := a.Interacts("customer-interacts-lb", "Customer accesses the platform via Load Balancer.", "customer", "load-balancer").
		Data("public", true)
	adminInteractsLB := a.Interacts("admin-interacts-lb", "Admin manages the platform via Load Balancer.", "admin", "load-balancer").
		Data("internal", true)

	// --- Wiring (Match original JSON order) ---

	lbToGateway1 := n.LB.ConnectTo(n.Gateways[0], "Load Balancer distributes traffic to Gateway Instance 1.").
		WithID("lb-connects-gateway-1").Via("lb-to-gateway", "gateway-1-http").Is("internal").Encrypted(true)

	lbToGateway2 := n.LB.ConnectTo(n.Gateways[1], "Load Balancer distributes traffic to Gateway Instance 2.").
		WithID("lb-connects-gateway-2").Via("lb-to-gateway", "gateway-2-http").Is("internal").Encrypted(true)

	gateway1ToOrder := n.Gateways[0].ConnectTo(n.OrderSvc, "Gateway 1 forwards requests to Order Service.").
		WithID("gateway-1-connects-order").Via("order-client-1", "order-api").Is("internal").Encrypted(true)

	n.Gateways[1].ConnectTo(n.OrderSvc, "Gateway 2 forwards requests to Order Service.").
		WithID("gateway-2-connects-order").Via("order-client-2", "order-api").Is("internal").Encrypted(true)

	n.Gateways[0].ConnectTo(n.InventorySvc, "Gateway 1 forwards requests to Inventory Service.").
		WithID("gateway-1-connects-inventory").Via("inventory-client-1", "inventory-api").Is("internal").Encrypted(true)

	gateway2ToInventory := n.Gateways[1].ConnectTo(n.InventorySvc, "Gateway 2 forwards requests to Inventory Service.").
		WithID("gateway-2-connects-inventory").Via("inventory-client-2", "inventory-api").Is("internal").Encrypted(true)

	n.OrderSvc.ConnectTo(n.OrderPrimary, "Order Service persists data to Primary Order Database.").
		WithID("order-connects-primary-db").
		Via("order-db-write-client", "order-sql-primary").Is("confidential").Encrypted(true).Tag("monitoring", true)

	orderToQueue := n.OrderSvc.ConnectTo(n.OrderQueue, "Order Service publishes payment task to queue.").
		WithID("order-publishes-to-queue").Via("payment-publisher", "").Is("internal").Encrypted(true).Protocol("AMQP")

	queueToPayment := n.OrderQueue.ConnectTo(n.PaymentSvc, "Payment Service consumes payment task from queue.").
		WithID("payment-subscribes-to-queue").
		Via("", "payment-consumer").Is("internal").Encrypted(true).Protocol("AMQP")

	n.OrderSvc.ConnectTo(n.OrderReplica, "Order Service reads data from Replica Order Database.").
		WithID("order-connects-replica-db").
		Via("order-db-read-client", "order-sql-replica").Is("confidential").Encrypted(true).Tag("monitoring", true)

	n.OrderSvc.ConnectTo(n.InventorySvc, "Order Service checks/reserves stock in Inventory Service.").
		WithID("order-connects-inventory").
		Via("order-inventory-client", "inventory-api").Is("internal").Encrypted(true).
		Tag("monitoring", true).Tag("circuit-breaker", true).Tag("latency-sla", "< 100ms")

	inventoryToDB := n.InventorySvc.ConnectTo(n.InventoryDB, "Inventory Service manages stock in Inventory Database.").
		WithID("inventory-connects-db").
		Via("inventory-db-client", "inventory-sql").Is("internal").Encrypted(true).Tag("monitoring", true)

	// --- Flows ---

	a.DefineFlow("order-processing-flow", "Customer Order Processing", "End-to-end flow from customer placing an order to payment confirmation").
		MetaMap(map[string]any{
			"business-impact":        "Customers cannot complete purchases - direct revenue loss",
			"degraded-behavior":      "Orders queue in message broker; processed when service recovers",
			"customer-communication": "Display 'Order processing delayed' message",
			"sla":                    "99.9% availability, 30s p99 latency",
		}).
		Steps(
			StepSpec{ID: customerInteractsLB.UniqueID, Desc: "Customer submits order via Load Balancer"},
			StepSpec{ID: lbToGateway1.GetID(), Desc: "LB routes to Gateway 1"},
			StepSpec{ID: gateway1ToOrder.GetID(), Desc: "Gateway 1 routes to Order Service"},
			StepSpec{ID: orderToQueue.GetID(), Desc: "Order Service publishes payment task"},
			StepSpec{ID: queueToPayment.GetID(), Desc: "Payment Service processes task from queue"},
		)

	a.DefineFlow("inventory-check-flow", "Inventory Stock Check", "Admin checks and updates inventory stock levels").
		MetaMap(map[string]any{
			"business-impact":        "Stock levels may be inaccurate - risk of overselling",
			"degraded-behavior":      "Fall back to cached inventory; flag orders for manual review",
			"customer-communication": "Display 'Stock availability being confirmed'",
			"sla":                    "99.5% availability, 500ms p99 latency",
		}).
		Steps(
			StepSpec{ID: adminInteractsLB.UniqueID, Desc: "Admin requests inventory status via LB"},
			StepSpec{ID: lbToGateway2.GetID(), Desc: "LB routes to Gateway 2"},
			StepSpec{ID: gateway2ToInventory.GetID(), Desc: "Gateway 2 routes to inventory service"},
			StepSpec{ID: inventoryToDB.GetID(), Desc: "Query current stock levels"},
			StepSpec{ID: inventoryToDB.GetID(), Desc: "Return stock data", Dir: "destination-to-source"},
		)

	a.ComposedOf("broker-composition", "Message broker contains the order queue.", "message-broker", []string{"order-queue"}).
		Data("internal", false)
	a.ComposedOf("order-db-composition", "Primary and replica databases form the order database cluster.", "order-database-cluster", []string{"order-database-primary", "order-database-replica"}).
		Data("confidential", true)
	a.ComposedOf("ecommerce-system-composition", "The E-Commerce Platform comprises its core services and databases.", "ecommerce-system",
		[]string{
			"load-balancer",
			"api-gateway-1",
			"api-gateway-2",
			"order-service",
			"inventory-service",
			"payment-service",
			"order-database-cluster",
			"inventory-db",
			"message-broker",
		}).
		Data("internal", false)
}
