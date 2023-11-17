package settings

import (
	"neodeliver.com/engine/graphql"
	"neodeliver.com/engine/rbac"
)

// register graphql queries
func Init(s *graphql.Builder) {
	// query user
	s.MongoQuery(User{}).Where(func(r rbac.RBAC) map[string]interface{} {
		return map[string]interface{}{
			"_id": r.UserID,
		}
	})

	// query contact_settings
	s.MongoQuery(ContactSettings{}).Where(func(r rbac.RBAC) map[string]interface{} {
		return map[string]interface{}{
			"_id": r.OrganizationID,
		}
	})

	// query smtp
	s.MongoQuery(SMTP{}).Where(func(r rbac.RBAC) map[string]interface{} {
		return map[string]interface{}{
			"_id": r.OrganizationID,
		}
	})

	s.AddMutationMethods(Mutation{})
}

type Mutation struct{}