package def

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"git.parallelcoin.io/dev/9/pkg/util"
)

// SaveConfig writes all the data in Cats the config file at the root of DataDir
func (app *App) SaveConfig() {
	if app == nil {
		return
	}
	datadir, ok := app.Cats["app"]["datadir"].Value.Get().(string)
	if !ok {
		return
	}
	configFile := util.CleanAndExpandPath(filepath.Join(datadir, "config"), "")
	if util.EnsureDir(configFile) {
	}
	fh, err := os.Create(configFile)
	if err != nil {
		panic(err)
	}
	j, e := json.MarshalIndent(app, "", "\t")
	if e != nil {
		panic(e)
	}
	_, err = fmt.Fprint(fh, string(j))
	if err != nil {
		panic(err)
	}
}

// MarshalJSON cherrypicks Cats for the values needed to correctly configure it
// and some extra information to make the JSON output friendly to human editors
func (r *App) MarshalJSON() ([]byte, error) {
	out := make(CatsJSON)
	for i, x := range r.Cats {
		out[i] = make(CatJSON)
		for j, y := range x {
			min, _ := y.Min.Get().(int)
			max, _ := y.Max.Get().(int)
			out[i][j] = Line{
				Value:   y.Value.Get(),
				Default: y.Default.Get(),
				Min:     min,
				Max:     max,
				Usage:   y.Usage,
			}
		}
	}
	return json.Marshal(out)
}

// UnmarshalJSON takes the cherrypicked JSON output of Marshal and puts it back into
// an App
func (r *App) UnmarshalJSON(data []byte) error {
	out := make(CatsJSON)
	e := json.Unmarshal(data, &out)
	if e != nil {
		return e
	}
	for i, x := range out {
		for j, y := range x {
			R := r.Cats[i][j]
			if y.Value != nil {
				switch R.Type {
				case "int", "port":
					y.Value = int(y.Value.(float64))
				case "duration":
					y.Value = time.Duration(int(y.Value.(float64)))
				case "stringslice":
					rt, ok := y.Value.([]string)
					ro := []string{}
					if ok {
						for _, z := range rt {
							R.Validate(R, z)
							ro = append(ro, z)
						}
						// R.Value.Put(ro)
					}
					// break
				case "float":
				}
			}
			R.Validate(R, y.Value)
			// R.Value.Put(y.Value)
		}
	}
	return nil
}

// RunAll triggers AppGenerators to configure an App
func (r *AppGenerators) RunAll(app *App) {
	for _, x := range *r {
		x(app)
	}
	return
}

// RunAll runs a collection of CatGenerators on a Cat
func (r *CatGenerators) RunAll(cat Cat) {
	for _, x := range *r {
		x(&cat)
	}
	return
}

func (r *RowGenerators) RunAll(row *Row) {
	for _, x := range *r {
		x(row)
	}
}

func (r *CommandGenerators) RunAll() *Command {
	c := &Command{}
	for _, x := range *r {
		x(c)
	}
	return c
}
