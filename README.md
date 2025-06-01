# satusehat-be-golang
Backend Go Lang Integration SATUSEHAT, always using staging / sandbox as your environment

# initialize mongoDB
use satusehat_mirror
db.credentials.insertOne({ client_id: "example_client_id", client_secret: "example_secret_key", token_url: "https://api-satusehat-stg.dto.kemkes.go.id/oauth2/v1/accesstoken?grant_type=client_credentials" })
show dbs

# setting connectionString on main.go
change this string on line client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
change db mongodb://localhost:27017 based on you host:port

# run main script
run in terminal "go run main.go"

# test your endpoint
run in terminal

Get Patient
curl --location 'http://localhost:8080/simrs/v1/patient/your-patient-id'

Get Practitioner
curl --location 'http://localhost:8080/simrs/v1/practitioner/your-practitioner-id'

# Endpoint Operation

POST Location
http://localhost:8080/simrs/v1/location/create

POST Encounter
http://localhost:8080/simrs/v1/encounter/create

PUT Encounter
http://localhost:8080/simrs/v1/encounter/update/your-encounter-id

PATCH Encounter
http://localhost:8080/simrs/v1/encounter/patch/your-encounter-id

GET Audit-Trail / Logs (Decoded)
http://localhost:8080/simrs/v1/audit-logs