package knowledge

import (
    "database/sql"
    "fmt"
    "log"
    "mcp-compose-memory/internal/models"
    "strings"

    "github.com/lib/pq"
)

type Manager struct {
    db *sql.DB
}

func NewManager(db *sql.DB) *Manager {
    return &Manager{db: db}
}

func (m *Manager) getEntityByName(tx *sql.Tx, name string) (*models.Entity, error) {
    var entity models.Entity
    err := tx.QueryRow("SELECT id, name, entity_type FROM entities WHERE name = $1", name).
        Scan(&entity.ID, &entity.Name, &entity.EntityType)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    return &entity, nil
}

func (m *Manager) getEntityObservations(tx *sql.Tx, entityID int) ([]string, error) {
    rows, err := tx.Query("SELECT content FROM observations WHERE entity_id = $1 ORDER BY created_at", entityID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var observations []string
    for rows.Next() {
        var content string
        if err := rows.Scan(&content); err != nil {
            return nil, err
        }
        observations = append(observations, content)
    }

    return observations, rows.Err()
}

func (m *Manager) CreateEntities(entities []models.Entity) ([]models.Entity, error) {
    tx, err := m.db.Begin()
    if err != nil {
        return nil, err
    }
    defer tx.Rollback()

    var newEntities []models.Entity

    for _, entity := range entities {
        existingEntity, err := m.getEntityByName(tx, entity.Name)
        if err != nil {
            return nil, err
        }

        if existingEntity == nil {
            var entityID int
            err := tx.QueryRow("INSERT INTO entities (name, entity_type) VALUES ($1, $2) RETURNING id",
                entity.Name, entity.EntityType).Scan(&entityID)
            if err != nil {
                return nil, err
            }

            for _, observation := range entity.Observations {
                _, err := tx.Exec("INSERT INTO observations (entity_id, content) VALUES ($1, $2)",
                    entityID, observation)
                if err != nil {
                    return nil, err
                }
            }

            newEntities = append(newEntities, entity)
        }
    }

    if err := tx.Commit(); err != nil {
        return nil, err
    }

    return newEntities, nil
}

func (m *Manager) CreateRelations(relations []models.Relation) ([]models.Relation, error) {
    tx, err := m.db.Begin()
    if err != nil {
        return nil, err
    }
    defer tx.Rollback()

    var newRelations []models.Relation

    for _, relation := range relations {
        fromEntity, err := m.getEntityByName(tx, relation.From)
        if err != nil {
            return nil, err
        }
        toEntity, err := m.getEntityByName(tx, relation.To)
        if err != nil {
            return nil, err
        }

        if fromEntity == nil || toEntity == nil {
            log.Printf("Skipping relation %s -> %s: entity not found", relation.From, relation.To)
            continue
        }

        var exists bool
        err = tx.QueryRow("SELECT EXISTS(SELECT 1 FROM relations WHERE from_entity_id = $1 AND to_entity_id = $2 AND relation_type = $3)",
            fromEntity.ID, toEntity.ID, relation.RelationType).Scan(&exists)
        if err != nil {
            return nil, err
        }

        if !exists {
            _, err := tx.Exec("INSERT INTO relations (from_entity_id, to_entity_id, relation_type) VALUES ($1, $2, $3)",
                fromEntity.ID, toEntity.ID, relation.RelationType)
            if err != nil {
                return nil, err
            }
            newRelations = append(newRelations, relation)
        }
    }

    if err := tx.Commit(); err != nil {
        return nil, err
    }

    return newRelations, nil
}

func (m *Manager) AddObservations(observations []struct {
    EntityName string   `json:"entityName"`
    Contents   []string `json:"contents"`
}) ([]struct {
    EntityName        string   `json:"entityName"`
    AddedObservations []string `json:"addedObservations"`
}, error) {
    tx, err := m.db.Begin()
    if err != nil {
        return nil, err
    }
    defer tx.Rollback()

    var results []struct {
        EntityName        string   `json:"entityName"`
        AddedObservations []string `json:"addedObservations"`
    }

    for _, obs := range observations {
        entity, err := m.getEntityByName(tx, obs.EntityName)
        if err != nil {
            return nil, err
        }
        if entity == nil {
            return nil, fmt.Errorf("entity with name %s not found", obs.EntityName)
        }

        existingObservations, err := m.getEntityObservations(tx, entity.ID)
        if err != nil {
            return nil, err
        }

        var addedObservations []string
        for _, content := range obs.Contents {
            found := false
            for _, existing := range existingObservations {
                if existing == content {
                    found = true
                    break
                }
            }

            if !found {
                _, err := tx.Exec("INSERT INTO observations (entity_id, content) VALUES ($1, $2)",
                    entity.ID, content)
                if err != nil {
                    return nil, err
                }
                addedObservations = append(addedObservations, content)
            }
        }

        results = append(results, struct {
            EntityName        string   `json:"entityName"`
            AddedObservations []string `json:"addedObservations"`
        }{
            EntityName:        obs.EntityName,
            AddedObservations: addedObservations,
        })
    }

    if err := tx.Commit(); err != nil {
        return nil, err
    }

    return results, nil
}

func (m *Manager) DeleteEntities(entityNames []string) error {
    tx, err := m.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    for _, name := range entityNames {
        _, err := tx.Exec("DELETE FROM entities WHERE name = $1", name)
        if err != nil {
            return err
        }
    }

    return tx.Commit()
}

func (m *Manager) DeleteObservations(deletions []struct {
    EntityName   string   `json:"entityName"`
    Observations []string `json:"observations"`
}) error {
    tx, err := m.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    for _, deletion := range deletions {
        entity, err := m.getEntityByName(tx, deletion.EntityName)
        if err != nil {
            return err
        }
        if entity != nil {
            for _, observation := range deletion.Observations {
                _, err := tx.Exec("DELETE FROM observations WHERE entity_id = $1 AND content = $2",
                    entity.ID, observation)
                if err != nil {
                    return err
                }
            }
        }
    }

    return tx.Commit()
}

func (m *Manager) DeleteRelations(relations []models.Relation) error {
    tx, err := m.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    for _, relation := range relations {
        fromEntity, err := m.getEntityByName(tx, relation.From)
        if err != nil {
            return err
        }
        toEntity, err := m.getEntityByName(tx, relation.To)
        if err != nil {
            return err
        }

        if fromEntity != nil && toEntity != nil {
            _, err := tx.Exec("DELETE FROM relations WHERE from_entity_id = $1 AND to_entity_id = $2 AND relation_type = $3",
                fromEntity.ID, toEntity.ID, relation.RelationType)
            if err != nil {
                return err
            }
        }
    }

    return tx.Commit()
}

func (m *Manager) ReadGraph() (*models.KnowledgeGraph, error) {
    // Get entities with observations
    rows, err := m.db.Query(`
        SELECT e.name, e.entity_type,
               COALESCE(array_agg(o.content ORDER BY o.created_at) FILTER (WHERE o.content IS NOT NULL), ARRAY[]::text[]) as observations
        FROM entities e
        LEFT JOIN observations o ON e.id = o.entity_id
        GROUP BY e.id, e.name, e.entity_type
        ORDER BY e.name
    `)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var entities []models.Entity
    for rows.Next() {
        var entity models.Entity
        var observations pq.StringArray

        err := rows.Scan(&entity.Name, &entity.EntityType, &observations)
        if err != nil {
            return nil, err
        }

        entity.Observations = []string(observations)
        entities = append(entities, entity)
    }

    // Get relations
    relationRows, err := m.db.Query(`
        SELECT ef.name as from_name, et.name as to_name, r.relation_type
        FROM relations r
        JOIN entities ef ON r.from_entity_id = ef.id
        JOIN entities et ON r.to_entity_id = et.id
        ORDER BY ef.name, et.name
    `)
    if err != nil {
        return nil, err
    }
    defer relationRows.Close()

    var relations []models.Relation
    for relationRows.Next() {
        var relation models.Relation
        err := relationRows.Scan(&relation.From, &relation.To, &relation.RelationType)
        if err != nil {
            return nil, err
        }
        relations = append(relations, relation)
    }

    return &models.KnowledgeGraph{
        Entities:  entities,
        Relations: relations,
    }, nil
}

func (m *Manager) SearchNodes(query string) (*models.KnowledgeGraph, error) {
    rows, err := m.db.Query(`
        SELECT DISTINCT e.name, e.entity_type,
               COALESCE(array_agg(o.content ORDER BY o.created_at) FILTER (WHERE o.content IS NOT NULL), ARRAY[]::text[]) as observations
        FROM entities e
        LEFT JOIN observations o ON e.id = o.entity_id
        WHERE e.name ILIKE $1
           OR e.entity_type ILIKE $1
           OR to_tsvector('english', e.name) @@ plainto_tsquery('english', $2)
           OR EXISTS (
             SELECT 1 FROM observations obs 
             WHERE obs.entity_id = e.id 
             AND (obs.content ILIKE $1 OR to_tsvector('english', obs.content) @@ plainto_tsquery('english', $2))
           )
        GROUP BY e.id, e.name, e.entity_type
        ORDER BY e.name
    `, "%"+query+"%", query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var entities []models.Entity
    var entityNames []string

    for rows.Next() {
        var entity models.Entity
        var observations pq.StringArray

        err := rows.Scan(&entity.Name, &entity.EntityType, &observations)
        if err != nil {
            return nil, err
        }

        entity.Observations = []string(observations)
        entities = append(entities, entity)
        entityNames = append(entityNames, entity.Name)
    }

    if len(entityNames) == 0 {
        return &models.KnowledgeGraph{Entities: []models.Entity{}, Relations: []models.Relation{}}, nil
    }

    // Get relations between found entities
    relationRows, err := m.db.Query(`
        SELECT ef.name as from_name, et.name as to_name, r.relation_type
        FROM relations r
        JOIN entities ef ON r.from_entity_id = ef.id
        JOIN entities et ON r.to_entity_id = et.id
        WHERE ef.name = ANY($1) AND et.name = ANY($1)
        ORDER BY ef.name, et.name
    `, pq.Array(entityNames))
    if err != nil {
        return nil, err
    }
    defer relationRows.Close()

    var relations []models.Relation
    for relationRows.Next() {
        var relation models.Relation
        err := relationRows.Scan(&relation.From, &relation.To, &relation.RelationType)
        if err != nil {
            return nil, err
        }
        relations = append(relations, relation)
    }

    return &models.KnowledgeGraph{
        Entities:  entities,
        Relations: relations,
    }, nil
}

func (m *Manager) OpenNodes(names []string) (*models.KnowledgeGraph, error) {
    if len(names) == 0 {
        return &models.KnowledgeGraph{Entities: []models.Entity{}, Relations: []models.Relation{}}, nil
    }

    rows, err := m.db.Query(`
        SELECT e.name, e.entity_type,
               COALESCE(array_agg(o.content ORDER BY o.created_at) FILTER (WHERE o.content IS NOT NULL), ARRAY[]::text[]) as observations
        FROM entities e
        LEFT JOIN observations o ON e.id = o.entity_id
        WHERE e.name = ANY($1)
        GROUP BY e.id, e.name, e.entity_type
        ORDER BY e.name
    `, pq.Array(names))
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var entities []models.Entity
    for rows.Next() {
        var entity models.Entity
        var observations pq.StringArray

        err := rows.Scan(&entity.Name, &entity.EntityType, &observations)
        if err != nil {
            return nil, err
        }

        entity.Observations = []string(observations)
        entities = append(entities, entity)
    }

    relationRows, err := m.db.Query(`
        SELECT ef.name as from_name, et.name as to_name, r.relation_type
        FROM relations r
        JOIN entities ef ON r.from_entity_id = ef.id
        JOIN entities et ON r.to_entity_id = et.id
        WHERE ef.name = ANY($1) AND et.name = ANY($1)
        ORDER BY ef.name, et.name
    `, pq.Array(names))
    if err != nil {
        return nil, err
    }
    defer relationRows.Close()

    var relations []models.Relation
    for relationRows.Next() {
        var relation models.Relation
        err := relationRows.Scan(&relation.From, &relation.To, &relation.RelationType)
        if err != nil {
            return nil, err
        }
        relations = append(relations, relation)
    }

    return &models.KnowledgeGraph{
        Entities:  entities,
        Relations: relations,
    }, nil
}
