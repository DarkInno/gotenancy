module github.com/DarkInno/saas/examples/ent

go 1.23.0

require (
	entgo.io/ent v0.14.1
	github.com/DarkInno/saas v0.3.0
	github.com/DarkInno/saas/data/ent v0.3.0
)

require github.com/google/uuid v1.3.0 // indirect

replace github.com/DarkInno/saas => ../..

replace github.com/DarkInno/saas/data/ent => ../../data/ent
