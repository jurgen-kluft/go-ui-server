package editor

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"unicode"
)

//
// ============================================================
// TOKENIZER
// ============================================================
//

type tokenKind int

const (
	tEOF tokenKind = iota
	tIdent
	tNumber
	tString
	tSymbol
)

type token struct {
	kind tokenKind
	text string
}

type lexer struct {
	src []rune
	pos int
}

func newLexer(input string) *lexer {
	return &lexer{src: []rune(input)}
}

func (l *lexer) next() token {
	l.skip()

	if l.pos >= len(l.src) {
		return token{kind: tEOF}
	}

	ch := l.src[l.pos]

	// identifier
	if unicode.IsLetter(ch) || ch == '_' {
		start := l.pos
		for l.pos < len(l.src) &&
			(unicode.IsLetter(l.src[l.pos]) || unicode.IsDigit(l.src[l.pos]) || l.src[l.pos] == '_') {
			l.pos++
		}
		return token{tIdent, string(l.src[start:l.pos])}
	}

	// number
	if unicode.IsDigit(ch) {
		start := l.pos
		for l.pos < len(l.src) && unicode.IsDigit(l.src[l.pos]) {
			l.pos++
		}
		return token{tNumber, string(l.src[start:l.pos])}
	}

	// hex color literal
	if ch == '#' && l.pos+1 < len(l.src) && isHexDigit(l.src[l.pos+1]) {
		start := l.pos
		l.pos++
		for l.pos < len(l.src) && isHexDigit(l.src[l.pos]) {
			l.pos++
		}
		return token{tIdent, string(l.src[start:l.pos])}
	}

	// string
	if ch == '"' {
		l.pos++
		start := l.pos
		for l.pos < len(l.src) && l.src[l.pos] != '"' {
			l.pos++
		}
		txt := string(l.src[start:l.pos])
		if l.pos < len(l.src) {
			l.pos++
		}
		return token{tString, txt}
	}

	// symbol
	l.pos++
	return token{tSymbol, string(ch)}
}

func (l *lexer) skip() {
	for l.pos < len(l.src) {
		if unicode.IsSpace(l.src[l.pos]) {
			l.pos++
			continue
		}
		if l.src[l.pos] == '#' && (l.pos+1 >= len(l.src) || !isHexDigit(l.src[l.pos+1])) {
			for l.pos < len(l.src) && l.src[l.pos] != '\n' {
				l.pos++
			}
			continue
		}
		break
	}
}

func isHexDigit(ch rune) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

//
// ============================================================
// ERRORS
// ============================================================
//

type blockingErrors struct {
	msgs []string
}

func (b *blockingErrors) add(msg string) {
	b.msgs = append(b.msgs, msg)
}

func (b *blockingErrors) toError(prefix string) error {
	if len(b.msgs) == 0 {
		return nil
	}
	return errors.New(prefix + ": " + strings.Join(b.msgs, "; "))
}

func logParseError(b *blockingErrors, blocking bool, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Printf("ui_dsl_parser error: %s", msg)
	if blocking && b != nil {
		b.add(msg)
	}
}

//
// ============================================================
// EXPRESSIONS (AST ONLY)
// ============================================================
//

type expr interface{}

type exNum struct{ v int }
type exIdent struct{ name string }
type exBin struct {
	op string
	l  expr
	r  expr
}

//
// ============================================================
// PARSER
// ============================================================
//

type parser struct {
	lex *lexer
	tok token
	err *blockingErrors
}

func (p *parser) next() { p.tok = p.lex.next() }

func (p *parser) expect(k tokenKind, txt string) {
	if p.tok.kind != k || (txt != "" && p.tok.text != txt) {
		logParseError(p.err, true, "unexpected token: got=%q expectedKind=%d expectedText=%q", p.tok.text, k, txt)
		if p.tok.kind == tEOF {
			return
		}
	}
	p.next()
}

//
// ---------------------
// Expression parsing
// ---------------------
//

func (p *parser) parseExpr() expr {
	return p.parseAdd()
}

func (p *parser) parseAdd() expr {
	e := p.parseMul()
	for p.tok.text == "+" || p.tok.text == "-" {
		op := p.tok.text
		p.next()
		e = exBin{op, e, p.parseMul()}
	}
	return e
}

func (p *parser) parseMul() expr {
	e := p.parsePrim()
	for p.tok.text == "*" || p.tok.text == "/" {
		op := p.tok.text
		p.next()
		e = exBin{op, e, p.parsePrim()}
	}
	return e
}

func (p *parser) parsePrim() expr {
	if p.tok.kind == tNumber {
		v, _ := strconv.Atoi(p.tok.text)
		p.next()
		return exNum{v}
	}
	if p.tok.kind == tIdent {
		name := p.tok.text
		p.next()
		return exIdent{name}
	}
	if p.tok.text == "(" {
		p.next()
		e := p.parseExpr()
		p.expect(tSymbol, ")")
		return e
	}
	logParseError(p.err, true, "invalid expression near token: %q", p.tok.text)
	if p.tok.kind != tEOF {
		p.next()
	}
	return exNum{0}
}

//
// ============================================================
// AST
// ============================================================
//

type astProgram struct {
	screen     map[string]expr
	colors     map[string]string
	components map[string]astComponent
	pages      []astPage
}

type astComponent struct {
	name string
	cmds []astCmd
}

type astPage struct {
	name string
	cmds []astCmd
}

type astCmd struct {
	kind string
	args []interface{} // expr OR string
}

//
// ============================================================
// PARSE STRUCTURE
// ============================================================
//

func parse(input string) (*astProgram, error) {
	errBag := &blockingErrors{}
	p := &parser{lex: newLexer(input), err: errBag}
	p.next()

	prog := &astProgram{
		screen:     map[string]expr{},
		colors:     map[string]string{},
		components: map[string]astComponent{},
	}

	for p.tok.kind != tEOF {
		switch p.tok.text {

		case "screen":
			parseScreen(p, prog)

		case "colors":
			parseColors(p, prog)

		case "component":
			c := parseComponent(p)
			prog.components[c.name] = c

		case "page":
			prog.pages = append(prog.pages, parsePage(p))

		default:
			logParseError(errBag, false, "unknown top-level block ignored: %q", p.tok.text)
			if p.tok.kind == tEOF {
				break
			}
			p.next()
		}
	}

	return prog, errBag.toError("parse blocked")
}

func parseScreen(p *parser, prog *astProgram) {
	p.next()
	for p.tok.kind != tEOF && p.tok.text != "end" {
		key := p.tok.text
		p.next()
		prog.screen[key] = p.parseExpr()
	}
	if p.tok.kind == tEOF {
		logParseError(p.err, true, "screen block missing terminating 'end'")
		return
	}
	p.next()
}

func parseColors(p *parser, prog *astProgram) {
	p.next()
	for p.tok.kind != tEOF && p.tok.text != "end" {
		name := p.tok.text
		p.next()
		val := p.tok.text
		p.next()
		prog.colors[name] = val
	}
	if p.tok.kind == tEOF {
		logParseError(p.err, true, "colors block missing terminating 'end'")
		return
	}
	p.next()
}

func parsePage(p *parser) astPage {
	p.next()
	name := p.tok.text
	p.expect(tIdent, "")

	page := astPage{name: name}

	for p.tok.kind != tEOF && p.tok.text != "end" {
		page.cmds = append(page.cmds, parseCmd(p))
	}
	if p.tok.kind == tEOF {
		logParseError(p.err, true, "page %q missing terminating 'end'", name)
		return page
	}

	p.next()
	return page
}

func parseComponent(p *parser) astComponent {
	p.next()
	name := readQualifiedIdent(p)

	component := astComponent{name: name}

	for p.tok.kind != tEOF && p.tok.text != "end" {
		component.cmds = append(component.cmds, parseCmd(p))
	}
	if p.tok.kind == tEOF {
		logParseError(p.err, true, "component %q missing terminating 'end'", name)
		return component
	}

	p.next()
	return component
}

func parseCmd(p *parser) astCmd {
	cmd := astCmd{kind: normalizeCmdKind(p.tok.text)}
	p.next()

	switch cmd.kind {

	case "line":
		cmd.args = []interface{}{
			readIdent(p),
			p.parseExpr(), p.parseExpr(),
			p.parseExpr(), p.parseExpr(),
			readIdent(p),
		}

	case "rect":
		cmd.args = []interface{}{
			readIdent(p),
			p.parseExpr(), p.parseExpr(),
			p.parseExpr(), p.parseExpr(),
			readIdent(p),
		}

	case "circle":
		cmd.args = []interface{}{
			readIdent(p),
			p.parseExpr(), p.parseExpr(),
			p.parseExpr(),
			readIdent(p),
		}

	case "text":
		cmd.args = []interface{}{
			readIdent(p),
			p.parseExpr(), p.parseExpr(),
			readIdent(p),
			readString(p),
		}

	case "sprite":
		cmd.args = []interface{}{
			readIdent(p),
			readIdent(p),
			p.parseExpr(), p.parseExpr(),
			p.parseExpr(), p.parseExpr(),
		}

	case "component":
		cmd.args = []interface{}{
			readQualifiedIdent(p),
			p.parseExpr(),
			p.parseExpr(),
		}

	case "optionals":
		cmd.args = []interface{}{
			readIdent(p),
			p.parseExpr(),
			p.parseExpr(),
			parseOptionals(p),
		}

	default:
		logParseError(p.err, false, "unsupported command ignored: %q", cmd.kind)
		cmd.kind = "noop"
		cmd.args = nil
	}

	return cmd
}

func normalizeCmdKind(kind string) string {
	switch kind {
	case "cmdLine":
		return "line"
	case "cmdCircle":
		return "circle"
	case "cmdSprite":
		return "sprite"
	case "cmdOptionals":
		return "optionals"
	default:
		return kind
	}
}

func parseOptionals(p *parser) []string {
	options := []string{}

	for p.tok.kind != tEOF && p.tok.text != "end" {
		if p.tok.text != "option" {
			logParseError(p.err, true, "optionals expected 'option' or 'end', got %q", p.tok.text)
			if p.tok.kind == tEOF {
				break
			}
			p.next()
			continue
		}
		p.next()
		options = append(options, readQualifiedIdent(p))
	}

	if p.tok.kind == tEOF {
		logParseError(p.err, true, "optionals block missing terminating 'end'")
		return options
	}

	p.next()
	return options
}

func readIdent(p *parser) string {
	s := p.tok.text
	p.expect(tIdent, "")
	return s
}

func readString(p *parser) string {
	s := p.tok.text
	p.expect(tString, "")
	return s
}

func readQualifiedIdent(p *parser) string {
	name := readIdent(p)
	for p.tok.text == "." {
		p.next()
		name += "." + readIdent(p)
	}
	return name
}

//
// ============================================================
// RESOLUTION PASS (DETERMINISTIC)
// ============================================================
//

type context struct {
	constants map[string]int
	colors    map[string]int
}

func resolve(prog *astProgram) (*context, error) {
	errBag := &blockingErrors{}
	ctx := &context{
		constants: map[string]int{},
		colors:    map[string]int{},
	}

	// Resolve screen constants deterministically, even when declarations
	// depend on symbols declared later in the block.
	pending := map[string]expr{}
	for k, v := range prog.screen {
		pending[k] = v
	}

	for len(pending) > 0 {
		progress := false

		for k, v := range pending {
			value, ok := evalKnown(v, ctx.constants)
			if !ok {
				continue
			}
			ctx.constants[k] = value
			delete(pending, k)
			progress = true
		}

		if !progress {
			for k := range pending {
				logParseError(errBag, true, "unresolved screen symbol: %s", k)
				ctx.constants[k] = 0
				delete(pending, k)
			}
		}
	}

	for k, hex := range prog.colors {
		ctx.colors[k] = parseHex(hex)
	}

	return ctx, errBag.toError("resolve blocked")
}

func eval(e expr, ctx *context, errBag *blockingErrors) int {
	switch v := e.(type) {

	case exNum:
		return v.v

	case exIdent:
		val, ok := ctx.constants[v.name]
		if !ok {
			logParseError(errBag, false, "unknown symbol: %s", v.name)
			return 0
		}
		return val

	case exBin:
		l := eval(v.l, ctx, errBag)
		r := eval(v.r, ctx, errBag)

		switch v.op {
		case "+":
			return l + r
		case "-":
			return l - r
		case "*":
			return l * r
		case "/":
			if r == 0 {
				logParseError(errBag, false, "division by zero")
				return 0
			}
			return l / r
		}
	}

	logParseError(errBag, false, "invalid expression node during evaluation")
	return 0
}

func evalKnown(e expr, known map[string]int) (int, bool) {
	switch v := e.(type) {

	case exNum:
		return v.v, true

	case exIdent:
		val, ok := known[v.name]
		return val, ok

	case exBin:
		l, ok := evalKnown(v.l, known)
		if !ok {
			return 0, false
		}
		r, ok := evalKnown(v.r, known)
		if !ok {
			return 0, false
		}

		switch v.op {
		case "+":
			return l + r, true
		case "-":
			return l - r, true
		case "*":
			return l * r, true
		case "/":
			if r == 0 {
				logParseError(nil, false, "division by zero")
				return 0, true
			}
			return l / r, true
		}
	}

	logParseError(nil, false, "invalid expression node during known-value evaluation")
	return 0, false
}

//
// ============================================================
// IR GENERATION
// ============================================================
//

type cmdRect struct {
	mode           int
	x0, y0, x1, y1 int
	color          int
}

type cmdLine struct {
	mode           int
	x0, y0, x1, y1 int
	color          int
}

type cmdCircle struct {
	mode   int
	x, y   int
	radius int
	color  int
}

type cmdText struct {
	x, y  int
	color int
	text  string
	font  string
}

type cmdSprite struct {
	effect int
	sprite string
	x, y   int
	w, h   int
}

type cmdComponent struct {
	name string
	x, y int
}

type cmdOptionals struct {
	promptID string
	x, y     int
	options  []string
}

type irCmdRef struct {
	kind  string
	index int
}

type irPage struct {
	name       string
	lines      []cmdLine
	rects      []cmdRect
	circles    []cmdCircle
	texts      []cmdText
	sprites    []cmdSprite
	components []cmdComponent
	optionals  []cmdOptionals
	order      []irCmdRef
}

func buildIR(prog *astProgram, ctx *context) ([]irPage, error) {
	errBag := &blockingErrors{}
	out := []irPage{}

	for _, p := range prog.pages {
		ip := irPage{name: p.name}

		for _, cmd := range p.cmds {
			appendIRCmd(&ip, cmd, prog, ctx, 0, 0, errBag)
		}

		out = append(out, ip)
	}

	return out, errBag.toError("ir build blocked")
}

func appendIRCmd(ip *irPage, cmd astCmd, prog *astProgram, ctx *context, ox int, oy int, errBag *blockingErrors) {
	defer func() {
		if r := recover(); r != nil {
			logParseError(errBag, true, "internal IR error while processing command %q: %v", cmd.kind, r)
		}
	}()

	switch cmd.kind {

	case "line":
		ip.lines = append(ip.lines, cmdLine{
			mode:  modeToInt(cmd.args[0].(string)),
			x0:    eval(cmd.args[1].(expr), ctx, errBag) + ox,
			y0:    eval(cmd.args[2].(expr), ctx, errBag) + oy,
			x1:    eval(cmd.args[3].(expr), ctx, errBag) + ox,
			y1:    eval(cmd.args[4].(expr), ctx, errBag) + oy,
			color: ctx.colors[cmd.args[5].(string)],
		})
		ip.order = append(ip.order, irCmdRef{kind: "line", index: len(ip.lines) - 1})

	case "rect":
		ip.rects = append(ip.rects, cmdRect{
			mode:  modeToInt(cmd.args[0].(string)),
			x0:    eval(cmd.args[1].(expr), ctx, errBag) + ox,
			y0:    eval(cmd.args[2].(expr), ctx, errBag) + oy,
			x1:    eval(cmd.args[3].(expr), ctx, errBag) + ox,
			y1:    eval(cmd.args[4].(expr), ctx, errBag) + oy,
			color: ctx.colors[cmd.args[5].(string)],
		})
		ip.order = append(ip.order, irCmdRef{kind: "rect", index: len(ip.rects) - 1})

	case "circle":
		ip.circles = append(ip.circles, cmdCircle{
			mode:   modeToInt(cmd.args[0].(string)),
			x:      eval(cmd.args[1].(expr), ctx, errBag) + ox,
			y:      eval(cmd.args[2].(expr), ctx, errBag) + oy,
			radius: eval(cmd.args[3].(expr), ctx, errBag),
			color:  ctx.colors[cmd.args[4].(string)],
		})
		ip.order = append(ip.order, irCmdRef{kind: "circle", index: len(ip.circles) - 1})

	case "text":
		ip.texts = append(ip.texts, cmdText{
			font:  cmd.args[0].(string),
			x:     eval(cmd.args[1].(expr), ctx, errBag) + ox,
			y:     eval(cmd.args[2].(expr), ctx, errBag) + oy,
			color: ctx.colors[cmd.args[3].(string)],
			text:  cmd.args[4].(string),
		})
		ip.order = append(ip.order, irCmdRef{kind: "text", index: len(ip.texts) - 1})

	case "sprite":
		ip.sprites = append(ip.sprites, cmdSprite{
			effect: effectToInt(cmd.args[0].(string)),
			sprite: cmd.args[1].(string),
			x:      eval(cmd.args[2].(expr), ctx, errBag) + ox,
			y:      eval(cmd.args[3].(expr), ctx, errBag) + oy,
			w:      eval(cmd.args[4].(expr), ctx, errBag),
			h:      eval(cmd.args[5].(expr), ctx, errBag),
		})
		ip.order = append(ip.order, irCmdRef{kind: "sprite", index: len(ip.sprites) - 1})

	case "component":
		componentName := cmd.args[0].(string)
		nextOX := ox + eval(cmd.args[1].(expr), ctx, errBag)
		nextOY := oy + eval(cmd.args[2].(expr), ctx, errBag)
		if _, ok := prog.components[componentName]; !ok {
			logParseError(errBag, false, "unknown component: %s", componentName)
		}
		ip.components = append(ip.components, cmdComponent{name: componentName, x: nextOX, y: nextOY})
		ip.order = append(ip.order, irCmdRef{kind: "component", index: len(ip.components) - 1})

	case "optionals":
		promptID := cmd.args[0].(string)
		nextX := ox + eval(cmd.args[1].(expr), ctx, errBag)
		nextY := oy + eval(cmd.args[2].(expr), ctx, errBag)
		options := cmd.args[3].([]string)
		for _, option := range options {
			if _, ok := prog.components[option]; !ok {
				logParseError(errBag, false, "optionals, option references unknown component: %s", option)
			}
		}
		ip.optionals = append(ip.optionals, cmdOptionals{
			promptID: promptID,
			x:        nextX,
			y:        nextY,
			options:  append([]string{}, options...),
		})
		ip.order = append(ip.order, irCmdRef{kind: "optionals", index: len(ip.optionals) - 1})

	case "noop":
		return

	default:
		logParseError(errBag, false, "unsupported IR command ignored: %s", cmd.kind)
	}
}

//
// ============================================================
// HELPERS
// ============================================================
//

func parseHex(s string) int {
	s = strings.TrimPrefix(s, "#")
	v, _ := strconv.ParseUint(s, 16, 32)
	return int(v)
}

func modeToInt(s string) int {
	switch s {
	case "fill":
		return 1
	case "stroke":
		return 2
	case "frame":
		return 3
	}
	return 0
}

func effectToInt(s string) int {
	switch s {
	case "normal":
		return 1
	case "invert":
		return 2
	case "greyscale":
		return 3
	case "mirror":
		return 4
	}
	return 0
}
