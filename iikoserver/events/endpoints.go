package events

// The event journal. All XML. The whole surface needs licence module 2200
// (API_EVENTS) and right B_VTJ on the authenticating user.
const (
	EndpointEvents         = "/api/events"          // XML, <eventsList>; GET by window or revision, POST to filter
	EndpointEventsAdd      = "/api/events/add"      // XML, <eventsList> in and out; writes
	EndpointEventsMetadata = "/api/events/metadata" // XML, <groupsList>; GET all, POST to filter
	EndpointEventSessions  = "/api/events/sessions" // XML, no documented shape
)
