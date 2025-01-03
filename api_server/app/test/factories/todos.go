package factories

import (
	models "app/models/generated"

	"github.com/Pallinder/go-randomdata"
	"github.com/bluele/factory-go/factory"
	"github.com/volatiletech/null/v8"
)

var TodoFactory = factory.NewFactory(
	&models.Todo{
		Title: randomdata.StringSample(),
		Content: null.String{String: randomdata.StringSample(), Valid: true},
	},
)
