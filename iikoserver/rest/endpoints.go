package rest

// Session paths. Both auth and logout are tagged POST in the docs markup but
// every real client uses GET and it works.
const (
	EndpointAuth    = "/api/auth"         // text/plain token
	EndpointLogout  = "/api/logout"       // text/plain, releases the licence seat
	EndpointLicence = "/api/licence/info" // text
)
