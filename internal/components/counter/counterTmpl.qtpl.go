// Code generated by qtc from "counterTmpl.qtpl". DO NOT EDIT.
// See https://github.com/valyala/quicktemplate for details.

//line counterTmpl.qtpl:1
package counter

//line counterTmpl.qtpl:1
import (
	qtio422016 "io"

	qt422016 "github.com/valyala/quicktemplate"
)

//line counterTmpl.qtpl:1
var (
	_ = qtio422016.Copy
	_ = qt422016.AcquireByteBuffer
)

//line counterTmpl.qtpl:1
func StreamCounterTemplate(qw422016 *qt422016.Writer, id string, count int) {
//line counterTmpl.qtpl:1
	qw422016.N().S(`
	<button id="`)
//line counterTmpl.qtpl:2
	qw422016.E().S(id)
//line counterTmpl.qtpl:2
	qw422016.N().S(`">Click me! Count: `)
//line counterTmpl.qtpl:2
	qw422016.N().D(count)
//line counterTmpl.qtpl:2
	qw422016.N().S(`</button>
`)
//line counterTmpl.qtpl:3
}

//line counterTmpl.qtpl:3
func WriteCounterTemplate(qq422016 qtio422016.Writer, id string, count int) {
//line counterTmpl.qtpl:3
	qw422016 := qt422016.AcquireWriter(qq422016)
//line counterTmpl.qtpl:3
	StreamCounterTemplate(qw422016, id, count)
//line counterTmpl.qtpl:3
	qt422016.ReleaseWriter(qw422016)
//line counterTmpl.qtpl:3
}

//line counterTmpl.qtpl:3
func CounterTemplate(id string, count int) string {
//line counterTmpl.qtpl:3
	qb422016 := qt422016.AcquireByteBuffer()
//line counterTmpl.qtpl:3
	WriteCounterTemplate(qb422016, id, count)
//line counterTmpl.qtpl:3
	qs422016 := string(qb422016.B)
//line counterTmpl.qtpl:3
	qt422016.ReleaseByteBuffer(qb422016)
//line counterTmpl.qtpl:3
	return qs422016
//line counterTmpl.qtpl:3
}
