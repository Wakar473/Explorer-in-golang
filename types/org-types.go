package types

import (
	
)

// type Org struct {
// 	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
// 	OrgName   string             `json:"org_name" bson:"org_name"`
// 	OrgKey    string             `json:"org_key" bson:"org_key"`
// 	E         string             `json:"e" bson:"e"`
// 	S         string             `json:"s" bson:"s"`
// 	G         string             `json:"g" bson:"g"`
// 	Status    int                `json:"status" bson:"status"`
// 	CreatedAt int64              `json:"created_at" bson:"created_at"`
// 	UpdatedAt int64              `json:"updated_at" bson:"updated_at"`
// }

type OrgView struct {
	ID        string `json:"id"`
	OrgName   string `json:"org_name"`
	OrgKey    string `json:"org_key"`
	Status    int    `json:"status"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

type SingleOrg struct {
	ID string `json:"id"`
}
