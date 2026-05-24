package editor

import "testing"

func compileUI(t *testing.T, src string) []irPage {
	t.Helper()
	ast, err := parse(src)
	if err != nil {
		t.Fatalf("parse returned blocking error: %v", err)
	}
	ctx, err := resolve(ast)
	if err != nil {
		t.Fatalf("resolve returned blocking error: %v", err)
	}
	ir, err := buildIR(ast, ctx)
	if err != nil {
		t.Fatalf("buildIR returned blocking error: %v", err)
	}
	return ir
}

func TestCompilePipelineBasic(t *testing.T) {
	src := `
    screen
      WIDTH 800
      HEIGHT 600
      HALF_WIDTH WIDTH / 2
    end

    colors
      WHITE #FFFFFF
      RED #FF0000
    end

    page TestPage
      rect fill 0 0 WIDTH HEIGHT WHITE
      rect stroke 10 20 HALF_WIDTH 100 RED
      text Roboto 50 60 RED "Test String"
    end
    `

	ir := compileUI(t, src)

	if len(ir) != 1 {
		t.Fatalf("expected 1 page, got %d", len(ir))
	}

	p := ir[0]
	if p.name != "TestPage" {
		t.Fatalf("page name mismatch: %q", p.name)
	}

	if len(p.rects) != 2 {
		t.Fatalf("expected 2 rects, got %d", len(p.rects))
	}

	r0 := p.rects[0]
	if r0.x0 != 0 || r0.y0 != 0 || r0.x1 != 800 || r0.y1 != 600 {
		t.Fatalf("rect0 coords incorrect: %+v", r0)
	}
	if r0.mode != 1 {
		t.Fatalf("rect0 mode incorrect: %d", r0.mode)
	}
	if r0.color != 0xFFFFFF {
		t.Fatalf("rect0 color incorrect: %#x", r0.color)
	}

	r1 := p.rects[1]
	if r1.x1 != 400 {
		t.Fatalf("HALF_WIDTH not resolved correctly: %d", r1.x1)
	}
	if r1.mode != 2 {
		t.Fatalf("rect1 mode incorrect: %d", r1.mode)
	}
	if r1.color != 0xFF0000 {
		t.Fatalf("rect1 color incorrect: %#x", r1.color)
	}

	if len(p.texts) != 1 {
		t.Fatalf("expected 1 text, got %d", len(p.texts))
	}

	t0 := p.texts[0]
	if t0.font != "Roboto" {
		t.Fatalf("text font incorrect: %q", t0.font)
	}
	if t0.x != 50 || t0.y != 60 {
		t.Fatalf("text position incorrect: %+v", t0)
	}
	if t0.color != 0xFF0000 {
		t.Fatalf("text color incorrect: %#x", t0.color)
	}
	if t0.text != "Test String" {
		t.Fatalf("text content incorrect: %q", t0.text)
	}

	if len(p.order) != 3 {
		t.Fatalf("expected 3 ordered commands, got %d", len(p.order))
	}
	if p.order[0].kind != "rect" || p.order[1].kind != "rect" || p.order[2].kind != "text" {
		t.Fatalf("unexpected draw order: %+v", p.order)
	}
}

func TestCompilePipelineSupportsAllCommandsAndAliases(t *testing.T) {
	src := `
    screen
      WIDTH 320
      HEIGHT 240
      CENTER_X WIDTH / 2
      CENTER_Y HEIGHT / 2
      PAD (10 + 6) * 2
    end

    colors
      BLACK #000000
      GREEN #00FF00
      BLUE #0000FF
      WHITE #FFFFFF
    end

    page FullPage
      line stroke 0 1 WIDTH HEIGHT GREEN
      cmdLine fill PAD 2 CENTER_X HEIGHT BLUE
      circle frame CENTER_X CENTER_Y 25 WHITE
      cmdCircle stroke 10 20 PAD GREEN
      text MainFont PAD CENTER_Y WHITE "Status"
      sprite greyscale player_idle 12 34 56 78
      cmdSprite mirror icon_alert 90 91 32 16
    end
    `

	ir := compileUI(t, src)
	p := ir[0]

	if len(p.lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(p.lines))
	}
	if p.lines[0].mode != 2 || p.lines[0].x1 != 320 || p.lines[0].y1 != 240 || p.lines[0].color != 0x00FF00 {
		t.Fatalf("first line incorrect: %+v", p.lines[0])
	}
	if p.lines[1].mode != 1 || p.lines[1].x0 != 32 || p.lines[1].x1 != 160 || p.lines[1].color != 0x0000FF {
		t.Fatalf("second line incorrect: %+v", p.lines[1])
	}

	if len(p.circles) != 2 {
		t.Fatalf("expected 2 circles, got %d", len(p.circles))
	}
	if p.circles[0].mode != 3 || p.circles[0].x != 160 || p.circles[0].y != 120 || p.circles[0].radius != 25 || p.circles[0].color != 0xFFFFFF {
		t.Fatalf("first circle incorrect: %+v", p.circles[0])
	}
	if p.circles[1].mode != 2 || p.circles[1].radius != 32 || p.circles[1].color != 0x00FF00 {
		t.Fatalf("second circle incorrect: %+v", p.circles[1])
	}

	if len(p.texts) != 1 {
		t.Fatalf("expected 1 text, got %d", len(p.texts))
	}
	if p.texts[0].font != "MainFont" || p.texts[0].x != 32 || p.texts[0].y != 120 || p.texts[0].color != 0xFFFFFF || p.texts[0].text != "Status" {
		t.Fatalf("text command incorrect: %+v", p.texts[0])
	}

	if len(p.sprites) != 2 {
		t.Fatalf("expected 2 sprites, got %d", len(p.sprites))
	}
	if p.sprites[0].effect != 3 || p.sprites[0].sprite != "player_idle" || p.sprites[0].x != 12 || p.sprites[0].y != 34 || p.sprites[0].w != 56 || p.sprites[0].h != 78 {
		t.Fatalf("first sprite incorrect: %+v", p.sprites[0])
	}
	if p.sprites[1].effect != 4 || p.sprites[1].sprite != "icon_alert" || p.sprites[1].x != 90 || p.sprites[1].y != 91 || p.sprites[1].w != 32 || p.sprites[1].h != 16 {
		t.Fatalf("second sprite incorrect: %+v", p.sprites[1])
	}
}

func TestParseIgnoresCommentsButKeepsHexColors(t *testing.T) {
	src := `
    # leading comment
    screen
      WIDTH 64 # inline comment after expression
      HEIGHT 32
    end

    colors
      WHITE #FFFFFF
      BLACK #000000 # trailing comment after color
    end

    page CommentPage
      # comment before command
      rect fill 0 0 WIDTH HEIGHT WHITE
    end
    `

	ir := compileUI(t, src)
	p := ir[0]

	if len(p.rects) != 1 {
		t.Fatalf("expected 1 rect, got %d", len(p.rects))
	}
	if p.rects[0].x1 != 64 || p.rects[0].y1 != 32 {
		t.Fatalf("rect dimensions incorrect: %+v", p.rects[0])
	}
	if p.rects[0].color != 0xFFFFFF {
		t.Fatalf("hex color was not preserved: %#x", p.rects[0].color)
	}
}

func TestCompilePipelineComponentExpansion(t *testing.T) {
	src := `
    screen
      WIDTH 200
      HEIGHT 100
      HALF WIDTH / 2
    end

    colors
      WHITE #FFFFFF
      RED #FF0000
      GREEN #00FF00
    end

    component nice.button
      line stroke 0 0 WIDTH 0 GREEN
      rect fill 0 0 HALF HEIGHT WHITE
      sprite normal button_idle 20 30 40 50
    end

    page TestPage
	  component nice.button 7 9
    end
    `

	ast, err := parse(src)
	if err != nil {
		t.Fatalf("parse returned blocking error: %v", err)
	}
	if len(ast.components) != 1 {
		t.Fatalf("expected 1 component, got %d", len(ast.components))
	}
	if _, ok := ast.components["nice.button"]; !ok {
		t.Fatal("component nice.button not registered")
	}

	ctx, err := resolve(ast)
	if err != nil {
		t.Fatalf("resolve returned blocking error: %v", err)
	}
	ir, err := buildIR(ast, ctx)
	if err != nil {
		t.Fatalf("buildIR returned blocking error: %v", err)
	}
	if len(ir) != 1 {
		t.Fatalf("expected 1 page, got %d", len(ir))
	}

	p := ir[0]
	if len(p.lines) != 0 || len(p.rects) != 0 || len(p.sprites) != 0 {
		t.Fatalf("component should not be expanded into primitive commands: %+v", p)
	}
	if len(p.components) != 1 {
		t.Fatalf("expected 1 component command, got %d", len(p.components))
	}
	c0 := p.components[0]
	if c0.name != "nice.button" || c0.x != 7 || c0.y != 9 {
		t.Fatalf("component command incorrect: %+v", c0)
	}
	if len(p.order) != 1 || p.order[0].kind != "component" {
		t.Fatalf("component order incorrect: %+v", p.order)
	}
}

func TestCompilePipelinePreservesInterleavedDrawOrder(t *testing.T) {
	src := `
    screen
      WIDTH 100
      HEIGHT 60
    end

    colors
      WHITE #FFFFFF
      RED #FF0000
      GREEN #00FF00
    end

    component nice.button
      rect fill 1 1 10 10 WHITE
    end

    page OrderPage
      line stroke 0 0 10 10 RED
      component nice.button 3 4
      circle frame 50 20 5 GREEN
      text Main 2 3 WHITE "txt"
      rect fill 0 0 WIDTH HEIGHT WHITE
    end
    `

	p := compileUI(t, src)[0]

	if len(p.order) != 5 {
		t.Fatalf("expected 5 ordered commands, got %d", len(p.order))
	}
	wantKinds := []string{"line", "component", "circle", "text", "rect"}
	for i, want := range wantKinds {
		if p.order[i].kind != want {
			t.Fatalf("order[%d] = %q, want %q", i, p.order[i].kind, want)
		}
	}

	if len(p.components) != 1 || p.components[0].name != "nice.button" || p.components[0].x != 3 || p.components[0].y != 4 {
		t.Fatalf("component payload incorrect: %+v", p.components)
	}
}

func TestCompilePipelineAskComponents(t *testing.T) {
	src := `
    screen
      WIDTH 120
      HEIGHT 80
    end

    colors
      WHITE #FFFFFF
    end

    component nice.button
      rect fill 0 0 10 10 WHITE
    end

    component alert.card
      rect fill 0 0 5 5 WHITE
    end

    page AskPage
      line stroke 0 0 10 10 WHITE
      optionals pickMain 7 9
        option nice.button
        option alert.card
      end
      rect fill 0 0 WIDTH HEIGHT WHITE
    end
    `

	p := compileUI(t, src)[0]

	if len(p.optionals) != 1 {
		t.Fatalf("expected 1 optional command, got %d", len(p.optionals))
	}
	a0 := p.optionals[0]
	if a0.promptID != "pickMain" || a0.x != 7 || a0.y != 9 {
		t.Fatalf("optional position/prompt incorrect: %+v", a0)
	}
	if len(a0.options) != 2 || a0.options[0] != "nice.button" || a0.options[1] != "alert.card" {
		t.Fatalf("optional options incorrect: %+v", a0.options)
	}

	if len(p.order) != 3 {
		t.Fatalf("expected 3 ordered commands, got %d", len(p.order))
	}
	wantKinds := []string{"line", "optionals", "rect"}
	for i, want := range wantKinds {
		if p.order[i].kind != want {
			t.Fatalf("order[%d] = %q, want %q", i, p.order[i].kind, want)
		}
	}
}

func TestHelperMappings(t *testing.T) {
	modeTests := []struct {
		name string
		in   string
		want int
	}{
		{name: "fill", in: "fill", want: 1},
		{name: "stroke", in: "stroke", want: 2},
		{name: "frame", in: "frame", want: 3},
		{name: "unknown", in: "nope", want: 0},
	}

	for _, tc := range modeTests {
		if got := modeToInt(tc.in); got != tc.want {
			t.Fatalf("modeToInt(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}

	effectTests := []struct {
		name string
		in   string
		want int
	}{
		{name: "normal", in: "normal", want: 1},
		{name: "invert", in: "invert", want: 2},
		{name: "greyscale", in: "greyscale", want: 3},
		{name: "mirror", in: "mirror", want: 4},
		{name: "unknown", in: "nope", want: 0},
	}

	for _, tc := range effectTests {
		if got := effectToInt(tc.in); got != tc.want {
			t.Fatalf("effectToInt(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}

	if got := parseHex("#ABCDEF"); got != 0xABCDEF {
		t.Fatalf("parseHex returned %#x", got)
	}
	if got := parseHex("123456"); got != 0x123456 {
		t.Fatalf("parseHex without prefix returned %#x", got)
	}
	if !isHexDigit('F') || !isHexDigit('9') || isHexDigit('x') {
		t.Fatal("isHexDigit classification incorrect")
	}
}
