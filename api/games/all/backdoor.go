// Package all pulls every game into the build. Each game adds one file here
// (<id>.go) with a blank import of its package, whose init() registers it.
package all

import _ "github.com/nelsul/games-night/api/games/backdoor"
