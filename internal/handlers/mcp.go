package handlers

import (
    "encoding/json"
    "io"
    "log"
    "mcp-compose-memory/internal/knowledge"
    "mcp-compose-memory/internal/models"
    "net/http"
)

type MCPHandler struct {
    manager *knowledge.Manager
}

func NewMCPHandler(manager *knowledge.Manager) *MCPHandler {
    return &MCPHandler{manager: manager}
}

func (h *MCPHandler) HandleMCPRequest(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    body, err := io.ReadAll(r.Body)
    if err != nil {
        h.sendError(w, nil, -32700, "Parse error")
        return
    }

    log.Printf("Received HTTP request: %s", string(body))

    var request models.MCPRequest
    if err := json.Unmarshal(body, &request); err != nil {
        h.sendError(w, nil, -32700, "Parse error")
        return
    }

    switch request.Method {
    case "initialize":
        h.handleInitialize(w, &request)
    case "tools/list":
        h.handleToolsList(w, &request)
    case "tools/call":
        h.handleToolsCall(w, &request)
    default:
        h.sendError(w, request.ID, -32601, "Method not found")
    }
}

func (h *MCPHandler) handleInitialize(w http.ResponseWriter, request *models.MCPRequest) {
    response := models.MCPResponse{
        ID:      request.ID,
        JSONRPC: "2.0",
        Result: map[string]interface{}{
            "protocolVersion": "2025-03-26",
            "capabilities": map[string]interface{}{
                "tools": map[string]interface{}{},
            },
            "serverInfo": map[string]interface{}{
                "name":    "mcp-compose-memory",
                "version": "0.7.0",
            },
        },
    }
    h.sendResponse(w, &response)
}

func (h *MCPHandler) handleToolsList(w http.ResponseWriter, request *models.MCPRequest) {
    tools := []map[string]interface{}{
        {
            "name":        "create_entities",
            "description": "Create multiple new entities in the knowledge graph",
            "inputSchema": map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "entities": map[string]interface{}{
                        "type": "array",
                        "items": map[string]interface{}{
                            "type": "object",
                            "properties": map[string]interface{}{
                                "name":         map[string]interface{}{"type": "string", "description": "The name of the entity"},
                                "entityType":   map[string]interface{}{"type": "string", "description": "The type of the entity"},
                                "observations": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "An array of observation contents associated with the entity"},
                            },
                            "required": []string{"name", "entityType", "observations"},
                        },
                    },
                },
                "required": []string{"entities"},
            },
        },
        {
            "name":        "create_relations",
            "description": "Create multiple new relations between entities in the knowledge graph. Relations should be in active voice",
            "inputSchema": map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "relations": map[string]interface{}{
                        "type": "array",
                        "items": map[string]interface{}{
                            "type": "object",
                            "properties": map[string]interface{}{
                                "from":         map[string]interface{}{"type": "string", "description": "The name of the entity where the relation starts"},
                                "to":           map[string]interface{}{"type": "string", "description": "The name of the entity where the relation ends"},
                                "relationType": map[string]interface{}{"type": "string", "description": "The type of the relation"},
                            },
                            "required": []string{"from", "to", "relationType"},
                        },
                    },
                },
                "required": []string{"relations"},
            },
        },
        {
            "name":        "add_observations",
            "description": "Add new observations to existing entities in the knowledge graph",
            "inputSchema": map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "observations": map[string]interface{}{
                        "type": "array",
                        "items": map[string]interface{}{
                            "type": "object",
                            "properties": map[string]interface{}{
                                "entityName": map[string]interface{}{"type": "string", "description": "The name of the entity to add the observations to"},
                                "contents":   map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "An array of observation contents to add"},
                            },
                            "required": []string{"entityName", "contents"},
                        },
                    },
                },
                "required": []string{"observations"},
            },
        },
        {
            "name":        "delete_entities",
            "description": "Delete multiple entities and their associated relations from the knowledge graph",
            "inputSchema": map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "entityNames": map[string]interface{}{
                        "type":        "array",
                        "items":       map[string]interface{}{"type": "string"},
                        "description": "An array of entity names to delete",
                    },
                },
                "required": []string{"entityNames"},
            },
        },
        {
            "name":        "delete_observations",
            "description": "Delete specific observations from entities in the knowledge graph",
            "inputSchema": map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "deletions": map[string]interface{}{
                        "type": "array",
                        "items": map[string]interface{}{
                            "type": "object",
                            "properties": map[string]interface{}{
                                "entityName":   map[string]interface{}{"type": "string", "description": "The name of the entity containing the observations"},
                                "observations": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "An array of observations to delete"},
                            },
                            "required": []string{"entityName", "observations"},
                        },
                    },
                },
                "required": []string{"deletions"},
            },
        },
        {
            "name":        "delete_relations",
            "description": "Delete multiple relations from the knowledge graph",
            "inputSchema": map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "relations": map[string]interface{}{
                        "type": "array",
                        "items": map[string]interface{}{
                            "type": "object",
                            "properties": map[string]interface{}{
                                "from":         map[string]interface{}{"type": "string", "description": "The name of the entity where the relation starts"},
                                "to":           map[string]interface{}{"type": "string", "description": "The name of the entity where the relation ends"},
                                "relationType": map[string]interface{}{"type": "string", "description": "The type of the relation"},
                            },
                            "required": []string{"from", "to", "relationType"},
                        },
                        "description": "An array of relations to delete",
                    },
                },
                "required": []string{"relations"},
            },
        },
        {
            "name":        "read_graph",
            "description": "Read the entire knowledge graph",
            "inputSchema": map[string]interface{}{
                "type":       "object",
                "properties": map[string]interface{}{},
            },
        },
        {
            "name":        "search_nodes",
            "description": "Search for nodes in the knowledge graph based on a query",
            "inputSchema": map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "query": map[string]interface{}{"type": "string", "description": "The search query to match against entity names, types, and observation content"},
                },
                "required": []string{"query"},
            },
        },
        {
            "name":        "open_nodes",
            "description": "Open specific nodes in the knowledge graph by their names",
            "inputSchema": map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "names": map[string]interface{}{
                        "type":        "array",
                        "items":       map[string]interface{}{"type": "string"},
                        "description": "An array of entity names to retrieve",
                    },
                },
                "required": []string{"names"},
            },
        },
    }

    response := models.MCPResponse{
        ID:      request.ID,
        JSONRPC: "2.0",
        Result:  map[string]interface{}{"tools": tools},
    }
    h.sendResponse(w, &response)
}

func (h *MCPHandler) handleToolsCall(w http.ResponseWriter, request *models.MCPRequest) {
    paramsBytes, _ := json.Marshal(request.Params)
    var params models.ToolCallParams
    if err := json.Unmarshal(paramsBytes, &params); err != nil {
        h.sendError(w, request.ID, -32602, "Invalid params")
        return
    }

    var result interface{}
    var err error

    switch params.Name {
    case "create_entities":
        result, err = h.handleCreateEntities(params.Arguments)
    case "create_relations":
        result, err = h.handleCreateRelations(params.Arguments)
    case "add_observations":
        result, err = h.handleAddObservations(params.Arguments)
    case "delete_entities":
        result, err = h.handleDeleteEntities(params.Arguments)
    case "delete_observations":
        result, err = h.handleDeleteObservations(params.Arguments)
    case "delete_relations":
        result, err = h.handleDeleteRelations(params.Arguments)
    case "read_graph":
        result, err = h.handleReadGraph()
    case "search_nodes":
        result, err = h.handleSearchNodes(params.Arguments)
    case "open_nodes":
        result, err = h.handleOpenNodes(params.Arguments)
    default:
        h.sendError(w, request.ID, -32601, "Unknown tool: "+params.Name)
        return
    }

    if err != nil {
        log.Printf("Tool execution error: %v", err)
        h.sendError(w, request.ID, -32603, err.Error())
        return
    }

    response := models.MCPResponse{
        ID:      request.ID,
        JSONRPC: "2.0",
        Result:  result,
    }
    h.sendResponse(w, &response)
}

func (h *MCPHandler) handleCreateEntities(args map[string]interface{}) (interface{}, error) {
    argsBytes, _ := json.Marshal(args)
    var input models.CreateEntitiesInput
    if err := json.Unmarshal(argsBytes, &input); err != nil {
        return nil, err
    }

    entities, err := h.manager.CreateEntities(input.Entities)
    if err != nil {
        return nil, err
    }

    resultBytes, _ := json.Marshal(entities)
    return models.ToolResponse{
        Content: []models.ToolContent{{Type: "text", Text: string(resultBytes)}},
    }, nil
}

func (h *MCPHandler) handleCreateRelations(args map[string]interface{}) (interface{}, error) {
    argsBytes, _ := json.Marshal(args)
    var input models.CreateRelationsInput
    if err := json.Unmarshal(argsBytes, &input); err != nil {
        return nil, err
    }

    relations, err := h.manager.CreateRelations(input.Relations)
    if err != nil {
        return nil, err
    }

    resultBytes, _ := json.Marshal(relations)
    return models.ToolResponse{
        Content: []models.ToolContent{{Type: "text", Text: string(resultBytes)}},
    }, nil
}

func (h *MCPHandler) handleAddObservations(args map[string]interface{}) (interface{}, error) {
    argsBytes, _ := json.Marshal(args)
    var input models.AddObservationsInput
    if err := json.Unmarshal(argsBytes, &input); err != nil {
        return nil, err
    }

    results, err := h.manager.AddObservations(input.Observations)
    if err != nil {
        return nil, err
    }

    resultBytes, _ := json.Marshal(results)
    return models.ToolResponse{
        Content: []models.ToolContent{{Type: "text", Text: string(resultBytes)}},
    }, nil
}

func (h *MCPHandler) handleDeleteEntities(args map[string]interface{}) (interface{}, error) {
    argsBytes, _ := json.Marshal(args)
    var input models.DeleteEntitiesInput
    if err := json.Unmarshal(argsBytes, &input); err != nil {
        return nil, err
    }

    if err := h.manager.DeleteEntities(input.EntityNames); err != nil {
        return nil, err
    }

    return models.ToolResponse{
        Content: []models.ToolContent{{Type: "text", Text: "Entities deleted successfully"}},
    }, nil
}

func (h *MCPHandler) handleDeleteObservations(args map[string]interface{}) (interface{}, error) {
    argsBytes, _ := json.Marshal(args)
    var input models.DeleteObservationsInput
    if err := json.Unmarshal(argsBytes, &input); err != nil {
        return nil, err
    }

    if err := h.manager.DeleteObservations(input.Deletions); err != nil {
        return nil, err
    }

    return models.ToolResponse{
        Content: []models.ToolContent{{Type: "text", Text: "Observations deleted successfully"}},
    }, nil
}

func (h *MCPHandler) handleDeleteRelations(args map[string]interface{}) (interface{}, error) {
    argsBytes, _ := json.Marshal(args)
    var input models.DeleteRelationsInput
    if err := json.Unmarshal(argsBytes, &input); err != nil {
        return nil, err
    }

    if err := h.manager.DeleteRelations(input.Relations); err != nil {
        return nil, err
    }

    return models.ToolResponse{
        Content: []models.ToolContent{{Type: "text", Text: "Relations deleted successfully"}},
    }, nil
}

func (h *MCPHandler) handleReadGraph() (interface{}, error) {
    graph, err := h.manager.ReadGraph()
    if err != nil {
        return nil, err
    }

    resultBytes, _ := json.Marshal(graph)
    return models.ToolResponse{
        Content: []models.ToolContent{{Type: "text", Text: string(resultBytes)}},
    }, nil
}

func (h *MCPHandler) handleSearchNodes(args map[string]interface{}) (interface{}, error) {
    argsBytes, _ := json.Marshal(args)
    var input models.SearchNodesInput
    if err := json.Unmarshal(argsBytes, &input); err != nil {
        return nil, err
    }

    graph, err := h.manager.SearchNodes(input.Query)
    if err != nil {
        return nil, err
    }

    resultBytes, _ := json.Marshal(graph)
    return models.ToolResponse{
        Content: []models.ToolContent{{Type: "text", Text: string(resultBytes)}},
    }, nil
}

func (h *MCPHandler) handleOpenNodes(args map[string]interface{}) (interface{}, error) {
    argsBytes, _ := json.Marshal(args)
    var input models.OpenNodesInput
    if err := json.Unmarshal(argsBytes, &input); err != nil {
        return nil, err
    }

    graph, err := h.manager.OpenNodes(input.Names)
    if err != nil {
        return nil, err
    }

    resultBytes, _ := json.Marshal(graph)
    return models.ToolResponse{
        Content: []models.ToolContent{{Type: "text", Text: string(resultBytes)}},
    }, nil
}

func (h *MCPHandler) sendResponse(w http.ResponseWriter, response *models.MCPResponse) {
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}

func (h *MCPHandler) sendError(w http.ResponseWriter, id interface{}, code int, message string) {
    response := models.MCPResponse{
        ID:      id,
        JSONRPC: "2.0",
        Error: &models.MCPError{
            Code:    code,
            Message: message,
        },
    }
    w.WriteHeader(http.StatusOK) // JSON-RPC errors use 200 status
    json.NewEncoder(w).Encode(response)
}
