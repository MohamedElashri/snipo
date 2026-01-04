package demo

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/MohamedElashri/snipo/internal/models"
	"github.com/MohamedElashri/snipo/internal/services"
)

// Service handles demo mode functionality
type Service struct {
	db             *sql.DB
	snippetService *services.SnippetService
	logger         *slog.Logger
	resetInterval  time.Duration
	enabled        bool
}

// NewService creates a new demo service
func NewService(db *sql.DB, snippetService *services.SnippetService, logger *slog.Logger, resetInterval time.Duration, enabled bool) *Service {
	return &Service{
		db:             db,
		snippetService: snippetService,
		logger:         logger,
		resetInterval:  resetInterval,
		enabled:        enabled,
	}
}

// IsEnabled returns whether demo mode is enabled
func (s *Service) IsEnabled() bool {
	return s.enabled
}

// StartPeriodicReset starts the periodic database reset goroutine
func (s *Service) StartPeriodicReset(ctx context.Context) {
	if !s.enabled {
		return
	}

	s.logger.Warn("DEMO MODE ENABLED",
		"password", "demo",
		"reset_interval", s.resetInterval,
		"restrictions", "password changes and API key creation disabled")

	// Initial setup
	if err := s.ResetDatabase(ctx); err != nil {
		s.logger.Error("failed to initialize demo database", "error", err)
	}

	// Start periodic reset
	ticker := time.NewTicker(s.resetInterval)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				s.logger.Info("Demo mode: resetting database")
				if err := s.ResetDatabase(ctx); err != nil {
					s.logger.Error("failed to reset demo database", "error", err)
				} else {
					s.logger.Info("Demo mode: database reset complete")
				}
			}
		}
	}()
}

// ResetDatabase clears all data and creates fake snippets
func (s *Service) ResetDatabase(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Clear all data (preserve schema)
	tables := []string{
		"snippet_history",
		"snippet_files",
		"snippet_tags",
		"snippet_folders",
		"snippets",
		"tags",
		"folders",
		"api_tokens",
		"sessions",
	}

	for _, table := range tables {
		if _, err := tx.ExecContext(ctx, "DELETE FROM "+table); err != nil {
			s.logger.Warn("failed to clear table", "table", table, "error", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	// Create fake snippets
	return s.createFakeSnippets(ctx)
}

// createFakeSnippets generates 10 demo snippets with various capabilities
func (s *Service) createFakeSnippets(ctx context.Context) error {
	snippets := []models.SnippetInput{
		{
			Title:       "Hello World in Python",
			Description: "A simple Hello World program demonstrating basic Python syntax",
			Language:    "python",
			Content:     "#!/usr/bin/env python3\n\ndef main():\n    print(\"Hello, World!\")\n    print(\"Welcome to Snipo Demo!\")\n\nif __name__ == \"__main__\":\n    main()",
			Tags:        []string{"python", "basics", "demo"},
		},
		{
			Title:       "REST API Client",
			Description: "Multi-file example: A complete REST API client with configuration",
			Files: []models.SnippetFileInput{
				{
					Filename: "api_client.py",
					Content:  "import requests\nimport json\nfrom config import API_BASE_URL, API_KEY\n\nclass APIClient:\n    def __init__(self):\n        self.base_url = API_BASE_URL\n        self.headers = {'Authorization': f'Bearer {API_KEY}'}\n    \n    def get(self, endpoint):\n        response = requests.get(f\"{self.base_url}/{endpoint}\", headers=self.headers)\n        return response.json()\n    \n    def post(self, endpoint, data):\n        response = requests.post(f\"{self.base_url}/{endpoint}\", json=data, headers=self.headers)\n        return response.json()",
					Language: "python",
				},
				{
					Filename: "config.py",
					Content:  "# API Configuration\nAPI_BASE_URL = \"https://api.example.com/v1\"\nAPI_KEY = \"your-api-key-here\"\nTIMEOUT = 30\nRETRY_COUNT = 3",
					Language: "python",
				},
				{
					Filename: "main.py",
					Content:  "from api_client import APIClient\n\ndef main():\n    client = APIClient()\n    \n    # Fetch users\n    users = client.get('users')\n    print(f\"Found {len(users)} users\")\n    \n    # Create new user\n    new_user = client.post('users', {'name': 'John Doe', 'email': 'john@example.com'})\n    print(f\"Created user: {new_user['id']}\")\n\nif __name__ == '__main__':\n    main()",
					Language: "python",
				},
			},
			Tags: []string{"python", "api", "multi-file"},
		},
		{
			Title:       "Docker Compose Setup",
			Description: "Complete Docker setup for a web application with database",
			Language:    "yaml",
			Content:     "version: '3.8'\n\nservices:\n  web:\n    image: nginx:alpine\n    ports:\n      - \"80:80\"\n    volumes:\n      - ./html:/usr/share/nginx/html\n    depends_on:\n      - api\n  \n  api:\n    build: ./api\n    ports:\n      - \"3000:3000\"\n    environment:\n      - DATABASE_URL=postgresql://user:pass@db:5432/myapp\n    depends_on:\n      - db\n  \n  db:\n    image: postgres:14-alpine\n    environment:\n      - POSTGRES_USER=user\n      - POSTGRES_PASSWORD=pass\n      - POSTGRES_DB=myapp\n    volumes:\n      - postgres_data:/var/lib/postgresql/data\n\nvolumes:\n  postgres_data:",
			Tags:        []string{"docker", "devops", "yaml"},
		},
		{
			Title:       "React Custom Hook",
			Description: "A reusable React hook for fetching data with loading and error states",
			Language:    "javascript",
			Content:     "import { useState, useEffect } from 'react';\n\nexport function useFetch(url) {\n  const [data, setData] = useState(null);\n  const [loading, setLoading] = useState(true);\n  const [error, setError] = useState(null);\n\n  useEffect(() => {\n    const fetchData = async () => {\n      try {\n        setLoading(true);\n        const response = await fetch(url);\n        if (!response.ok) {\n          throw new Error(`HTTP error! status: ${response.status}`);\n        }\n        const json = await response.json();\n        setData(json);\n        setError(null);\n      } catch (e) {\n        setError(e.message);\n        setData(null);\n      } finally {\n        setLoading(false);\n      }\n    };\n\n    fetchData();\n  }, [url]);\n\n  return { data, loading, error };\n}",
			Tags:        []string{"react", "javascript", "hooks"},
		},
		{
			Title:       "SQL Query Examples",
			Description: "Common SQL queries for data analysis and reporting",
			Language:    "sql",
			Content:     "-- Get top 10 customers by revenue\nSELECT \n    c.customer_id,\n    c.name,\n    SUM(o.total_amount) as total_revenue\nFROM customers c\nJOIN orders o ON c.customer_id = o.customer_id\nWHERE o.order_date >= DATE_SUB(CURDATE(), INTERVAL 1 YEAR)\nGROUP BY c.customer_id, c.name\nORDER BY total_revenue DESC\nLIMIT 10;\n\n-- Find duplicate records\nSELECT email, COUNT(*) as count\nFROM users\nGROUP BY email\nHAVING COUNT(*) > 1;\n\n-- Monthly sales trend\nSELECT \n    DATE_FORMAT(order_date, '%Y-%m') as month,\n    COUNT(*) as order_count,\n    SUM(total_amount) as revenue\nFROM orders\nWHERE order_date >= DATE_SUB(CURDATE(), INTERVAL 12 MONTH)\nGROUP BY DATE_FORMAT(order_date, '%Y-%m')\nORDER BY month;",
			Tags:        []string{"sql", "database", "analytics"},
		},
		{
			Title:       "Bash Utility Scripts",
			Description: "Useful bash functions for system administration",
			Language:    "bash",
			Content:     "#!/bin/bash\n\n# Backup directory with timestamp\nbackup_dir() {\n    local source=\"$1\"\n    local dest=\"$2\"\n    local timestamp=$(date +%Y%m%d_%H%M%S)\n    tar -czf \"${dest}/backup_${timestamp}.tar.gz\" \"$source\"\n    echo \"Backup created: ${dest}/backup_${timestamp}.tar.gz\"\n}\n\n# Check disk usage and alert if above threshold\ncheck_disk_usage() {\n    local threshold=${1:-80}\n    df -h | awk -v thresh=$threshold '\n        NR>1 {\n            gsub(/%/, \"\", $5)\n            if ($5 > thresh) {\n                print \"WARNING: \" $6 \" is at \" $5 \"% capacity\"\n            }\n        }'\n}\n\n# Find large files\nfind_large_files() {\n    local dir=${1:-.}\n    local size=${2:-100M}\n    find \"$dir\" -type f -size +\"$size\" -exec ls -lh {} \\; | awk '{print $9 \": \" $5}'\n}\n\n# Usage examples\n# backup_dir /var/www /backups\n# check_disk_usage 90\n# find_large_files /home 500M",
			Tags:        []string{"bash", "linux", "sysadmin"},
		},
		{
			Title:       "Kubernetes Deployment",
			Description: "Example Kubernetes deployment with service and ingress",
			Language:    "yaml",
			Content:     "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: myapp\n  labels:\n    app: myapp\nspec:\n  replicas: 3\n  selector:\n    matchLabels:\n      app: myapp\n  template:\n    metadata:\n      labels:\n        app: myapp\n    spec:\n      containers:\n      - name: myapp\n        image: myapp:latest\n        ports:\n        - containerPort: 8080\n        env:\n        - name: DATABASE_URL\n          valueFrom:\n            secretKeyRef:\n              name: myapp-secrets\n              key: database-url\n        resources:\n          requests:\n            memory: \"128Mi\"\n            cpu: \"100m\"\n          limits:\n            memory: \"256Mi\"\n            cpu: \"200m\"\n---\napiVersion: v1\nkind: Service\nmetadata:\n  name: myapp-service\nspec:\n  selector:\n    app: myapp\n  ports:\n  - protocol: TCP\n    port: 80\n    targetPort: 8080\n  type: ClusterIP",
			Tags:        []string{"kubernetes", "devops", "yaml"},
		},
		{
			Title:       "Go HTTP Server",
			Description: "Simple HTTP server with middleware in Go",
			Language:    "go",
			Content:     "package main\n\nimport (\n\t\"encoding/json\"\n\t\"log\"\n\t\"net/http\"\n\t\"time\"\n)\n\ntype Response struct {\n\tMessage string    `json:\"message\"`\n\tTime    time.Time `json:\"time\"`\n}\n\nfunc loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {\n\treturn func(w http.ResponseWriter, r *http.Request) {\n\t\tstart := time.Now()\n\t\tlog.Printf(\"%s %s\", r.Method, r.URL.Path)\n\t\tnext(w, r)\n\t\tlog.Printf(\"Completed in %v\", time.Since(start))\n\t}\n}\n\nfunc handleRoot(w http.ResponseWriter, r *http.Request) {\n\tw.Header().Set(\"Content-Type\", \"application/json\")\n\tjson.NewEncoder(w).Encode(Response{\n\t\tMessage: \"Hello from Go!\",\n\t\tTime:    time.Now(),\n\t})\n}\n\nfunc main() {\n\thttp.HandleFunc(\"/\", loggingMiddleware(handleRoot))\n\tlog.Println(\"Server starting on :8080\")\n\tlog.Fatal(http.ListenAndServe(\":8080\", nil))\n}",
			Tags:        []string{"go", "http", "backend"},
		},
		{
			Title:       "CSS Grid Layout",
			Description: "Modern responsive grid layout with CSS Grid",
			Language:    "css",
			Content:     "/* Responsive Grid Layout */\n.container {\n  display: grid;\n  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));\n  gap: 2rem;\n  padding: 2rem;\n  max-width: 1200px;\n  margin: 0 auto;\n}\n\n.card {\n  background: white;\n  border-radius: 8px;\n  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);\n  padding: 1.5rem;\n  transition: transform 0.2s, box-shadow 0.2s;\n}\n\n.card:hover {\n  transform: translateY(-4px);\n  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);\n}\n\n/* Dashboard Layout */\n.dashboard {\n  display: grid;\n  grid-template-areas:\n    \"header header header\"\n    \"sidebar main main\"\n    \"sidebar footer footer\";\n  grid-template-columns: 250px 1fr 1fr;\n  grid-template-rows: auto 1fr auto;\n  min-height: 100vh;\n  gap: 1rem;\n}\n\n.header { grid-area: header; }\n.sidebar { grid-area: sidebar; }\n.main { grid-area: main; }\n.footer { grid-area: footer; }\n\n@media (max-width: 768px) {\n  .dashboard {\n    grid-template-areas:\n      \"header\"\n      \"main\"\n      \"sidebar\"\n      \"footer\";\n    grid-template-columns: 1fr;\n  }\n}",
			Tags:        []string{"css", "frontend", "responsive"},
		},
		{
			Title:       "Java REST API with Spring Boot",
			Description: "RESTful API example using Spring Boot with CRUD operations",
			Language:    "java",
			Content:     "package com.example.api;\n\nimport org.springframework.boot.SpringApplication;\nimport org.springframework.boot.autoconfigure.SpringBootApplication;\nimport org.springframework.web.bind.annotation.*;\nimport java.util.*;\n\n@SpringBootApplication\npublic class ApiApplication {\n    public static void main(String[] args) {\n        SpringApplication.run(ApiApplication.class, args);\n    }\n}\n\n@RestController\n@RequestMapping(\"/api/users\")\nclass UserController {\n    private List<User> users = new ArrayList<>();\n    private long nextId = 1;\n\n    @GetMapping\n    public List<User> getAllUsers() {\n        return users;\n    }\n\n    @GetMapping(\"/{id}\")\n    public User getUserById(@PathVariable Long id) {\n        return users.stream()\n            .filter(u -> u.getId().equals(id))\n            .findFirst()\n            .orElseThrow(() -> new RuntimeException(\"User not found\"));\n    }\n\n    @PostMapping\n    public User createUser(@RequestBody User user) {\n        user.setId(nextId++);\n        users.add(user);\n        return user;\n    }\n\n    @PutMapping(\"/{id}\")\n    public User updateUser(@PathVariable Long id, @RequestBody User updatedUser) {\n        User user = getUserById(id);\n        user.setName(updatedUser.getName());\n        user.setEmail(updatedUser.getEmail());\n        return user;\n    }\n\n    @DeleteMapping(\"/{id}\")\n    public void deleteUser(@PathVariable Long id) {\n        users.removeIf(u -> u.getId().equals(id));\n    }\n}\n\nclass User {\n    private Long id;\n    private String name;\n    private String email;\n\n    // Getters and setters\n    public Long getId() { return id; }\n    public void setId(Long id) { this.id = id; }\n    public String getName() { return name; }\n    public void setName(String name) { this.name = name; }\n    public String getEmail() { return email; }\n    public void setEmail(String email) { this.email = email; }\n}",
			Tags:        []string{"java", "spring-boot", "rest-api"},
		},
	}

	for _, input := range snippets {
		if _, err := s.snippetService.Create(ctx, &input); err != nil {
			s.logger.Warn("failed to create demo snippet", "title", input.Title, "error", err)
		}
	}

	s.logger.Info("created demo snippets", "count", len(snippets))
	return nil
}
