package models

import "time"

// Entity represents an entity in the knowledge graph
type Entity struct {
    ID           int       `json:"id" db:"id"`
    Name         string    `json:"name" db:"name"`
    EntityType   string    `json:"entityType" db:"entity_type"`
    Observations []string  `json:"observations"`
    CreatedAt    time.Time `json:"createdAt" db:"created_at"`
    UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
}

// Relation represents a relationship between entities
type Relation struct {
    ID           int       `json:"id" db:"id"`
    From         string    `json:"from"`
    To           string    `json:"to"`
    RelationType string    `json:"relationType" db:"relation_type"`
    CreatedAt    time.Time `json:"createdAt" db:"created_at"`
}

// Observation represents an observation about an entity
type Observation struct {
    ID        int       `json:"id" db:"id"`
    EntityID  int       `json:"entityId" db:"entity_id"`
    Content   string    `json:"content" db:"content"`
    CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

// KnowledgeGraph represents the entire graph structure
type KnowledgeGraph struct {
    Entities  []Entity   `json:"entities"`
    Relations []Relation `json:"relations"`
}

// MCP Protocol types
type MCPRequest struct {
    ID      interface{} `json:"id"`
    JSONRPC string      `json:"jsonrpc"`
    Method  string      `json:"method"`
    Params  interface{} `json:"params,omitempty"`
}

type MCPResponse struct {
    ID      interface{} `json:"id"`
    JSONRPC string      `json:"jsonrpc"`
    Result  interface{} `json:"result,omitempty"`
    Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

// Tool call parameters
type ToolCallParams struct {
    Name      string                 `json:"name"`
    Arguments map[string]interface{} `json:"arguments"`
}

// Tool response content
type ToolContent struct {
    Type string `json:"type"`
    Text string `json:"text"`
}

type ToolResponse struct {
    Content []ToolContent `json:"content"`
}

// Input schemas for tools
type CreateEntitiesInput struct {
    Entities []Entity `json:"entities"`
}

type CreateRelationsInput struct {
    Relations []Relation `json:"relations"`
}

type AddObservationsInput struct {
    Observations []struct {
        EntityName string   `json:"entityName"`
        Contents   []string `json:"contents"`
    } `json:"observations"`
}

type DeleteEntitiesInput struct {
    EntityNames []string `json:"entityNames"`
}

type DeleteObservationsInput struct {
    Deletions []struct {
        EntityName   string   `json:"entityName"`
        Observations []string `json:"observations"`
    } `json:"deletions"`
}

type DeleteRelationsInput struct {
    Relations []Relation `json:"relations"`
}

type SearchNodesInput struct {
    Query string `json:"query"`
}

type OpenNodesInput struct {
    Names []string `json:"names"`
}
