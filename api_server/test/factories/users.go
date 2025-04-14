package factories

import (
	models "app/models/generated"
	"log"

	"github.com/Pallinder/go-randomdata"
	"github.com/bluele/factory-go/factory"
	"golang.org/x/crypto/bcrypt"
)

var UserFactory = factory.NewFactory(
	&models.User{
		FirstName: randomdata.FirstName(randomdata.RandomGender),
		LastName: randomdata.LastName(),
		Email: randomdata.Email(),
		FrontIdentification: randomdata.StringSample(),
		BackIdentification: randomdata.StringSample(),
	},
).Attr("Password", func(args factory.Args) (interface{}, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to generate hash %v", err)
	}
	return string(hash), nil
})
