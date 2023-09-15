// Code generated by qtc from "appTmpl.qtpl". DO NOT EDIT.
// See https://github.com/valyala/quicktemplate for details.

//line appTmpl.qtpl:1
package app

//line appTmpl.qtpl:1
import (
	qtio422016 "io"

	qt422016 "github.com/valyala/quicktemplate"
)

//line appTmpl.qtpl:1
var (
	_ = qtio422016.Copy
	_ = qt422016.AcquireByteBuffer
)

//line appTmpl.qtpl:1
func StreamAppTemplate(qw422016 *qt422016.Writer, id string, name string, children []string, buttonID string) {
//line appTmpl.qtpl:1
	qw422016.N().S(`
  <div id="`)
//line appTmpl.qtpl:2
	qw422016.E().S(id)
//line appTmpl.qtpl:2
	qw422016.N().S(`" class="flex flex-col border">
    <h4 class="bg-red-100 text-center w-full p-3 font-bold">Hello, `)
//line appTmpl.qtpl:3
	qw422016.E().S(name)
//line appTmpl.qtpl:3
	qw422016.N().S(`!</h4>
    <div class="flex flex-col border">
      `)
//line appTmpl.qtpl:5
	for _, child := range children {
//line appTmpl.qtpl:5
		qw422016.N().S(`
        `)
//line appTmpl.qtpl:6
		qw422016.N().S(child)
//line appTmpl.qtpl:6
		qw422016.N().S(`
      `)
//line appTmpl.qtpl:7
	}
//line appTmpl.qtpl:7
	qw422016.N().S(`

      <button id="`)
//line appTmpl.qtpl:9
	qw422016.E().S(buttonID)
//line appTmpl.qtpl:9
	qw422016.N().S(`" class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded">
        Randomise
      </button>
    </div>
  </div>
`)
//line appTmpl.qtpl:14
}

//line appTmpl.qtpl:14
func WriteAppTemplate(qq422016 qtio422016.Writer, id string, name string, children []string, buttonID string) {
//line appTmpl.qtpl:14
	qw422016 := qt422016.AcquireWriter(qq422016)
//line appTmpl.qtpl:14
	StreamAppTemplate(qw422016, id, name, children, buttonID)
//line appTmpl.qtpl:14
	qt422016.ReleaseWriter(qw422016)
//line appTmpl.qtpl:14
}

//line appTmpl.qtpl:14
func AppTemplate(id string, name string, children []string, buttonID string) string {
//line appTmpl.qtpl:14
	qb422016 := qt422016.AcquireByteBuffer()
//line appTmpl.qtpl:14
	WriteAppTemplate(qb422016, id, name, children, buttonID)
//line appTmpl.qtpl:14
	qs422016 := string(qb422016.B)
//line appTmpl.qtpl:14
	qt422016.ReleaseByteBuffer(qb422016)
//line appTmpl.qtpl:14
	return qs422016
//line appTmpl.qtpl:14
}
