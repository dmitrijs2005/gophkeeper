// Package cli provides the interactive GophKeeper command-line client.
//
// It wires configuration, local storage, API services, and an interactive REPL
// that supports online/offline operation. Typical flow: prompt for credentials,
// start a background connectivity watcher, and execute user commands.
//
// Key features:
//   - Login / Logout (online with offline fallback)
//   - Add entries: notes, logins, credit cards, files
//   - List / Show entries
//   - Sync with the server
//
// The REPL is started via App.Root(ctx), which blocks until the user exits.
// See App, StartOnlineStatusWatcher, and runREPL for details.
package cli
