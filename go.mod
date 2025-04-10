module github.com/PDOK/gokoala

go 1.23.6

require (
	dario.cat/mergo v1.0.1
	github.com/PDOK/go-cloud-sqlite-vfs v0.3.0
	github.com/creasty/defaults v1.8.0
	github.com/docker/go-units v0.5.0
	github.com/elnormous/contenttype v1.0.4
	github.com/failsafe-go/failsafe-go v0.6.9
	github.com/getkin/kin-openapi v0.131.0
	github.com/go-chi/chi/v5 v5.2.0
	github.com/go-chi/cors v1.2.1
	github.com/go-playground/validator/v10 v10.24.0
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572
	github.com/goccy/go-json v0.10.5
	github.com/gomarkdown/markdown v0.0.0-20241205020045-f7e15b2f3e62
	github.com/google/uuid v1.6.0
	github.com/hashicorp/golang-lru/v2 v2.0.7
	github.com/jmoiron/sqlx v1.4.0
	github.com/mattn/go-sqlite3 v1.14.24
	github.com/nicksnyder/go-i18n/v2 v2.5.0
	github.com/qustavo/sqlhooks/v2 v2.1.0
	github.com/stretchr/testify v1.10.0
	github.com/twpayne/go-geom v1.6.0
	github.com/urfave/cli/v2 v2.27.5
	github.com/wk8/go-ordered-map/v2 v2.1.8
	github.com/writeas/go-strip-markdown/v2 v2.1.1
	go.uber.org/automaxprocs v1.6.0
	golang.org/x/sync v0.10.0
	golang.org/x/text v0.21.0
	gopkg.in/yaml.v3 v3.0.1
	schneider.vip/problem v1.9.1
)

// required until https://github.com/wk8/go-ordered-map/pull/45 is merged and released
replace github.com/wk8/go-ordered-map/v2 v2.1.8 => github.com/rkettelerij/go-ordered-map/v2 v2.2.1

require (
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.6 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.8 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/oasdiff/yaml v0.0.0-20250309154309-f31be36b4037 // indirect
	github.com/oasdiff/yaml3 v0.0.0-20250309153720-d2182401db90 // indirect
	github.com/perimeterx/marshmallow v1.1.5 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1 // indirect
	golang.org/x/crypto v0.32.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
)
