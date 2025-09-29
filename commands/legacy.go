package commands

import (
        "irc-client/internal/app"
)

// ConvertToNewRouter provides a way to use the new router with legacy ClientState
// This is a convenience function for gradual migration
func ConvertToNewRouter(state ClientState, userInput string) ([]app.Command, error) {
        // Convert old ClientState to new app.ClientState
        newState := app.ClientState{
                CurrentChannel: state.CurrentChannel,
                Nick:           state.Nick,
                Connected:      true, // Assume connected for legacy compatibility
        }
        
        // Create a router and route the command
        router := NewRouter()
        return router.Route(newState, userInput)
}