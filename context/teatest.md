[Charm](https://charm.land/)

We make the command

line glamorous.

- [Home](https://charm.land/)
- [Libs](https://charm.land/libs/)
- [Apps](https://charm.land/apps/)
- [Enterprise](https://charm.land/#enterprise)
- [Team](https://charm.land/team/)
- [Blog](https://charm.land/blog/)

- [172.0k](https://github.com/charmbracelet)

8 May 2023

# Writing Bubble Tea Tests

By [Carlos Becker](https://github.com/caarlos0)

Last week we launched our [`github.com/charmbracelet/x`](https://github.com/charmbracelet/x) repository,
which will contain experimental code that we’re not ready to promise any
compatibility guarantees on just yet.

The first module there is called [`teatest`](https://github.com/charmbracelet/x/tree/main/exp/teatest).
It is a library that aims to help you test your [Bubble Tea](https://github.com/charmbracelet/bubbletea) apps.

You can assert the entire output of your program, parts of it, and/or its
internal `tea.Model` state.

In this post we’ll add tests to an existing app using current’s `teatest`
version API.

## The app

Our example app is a simple `sleep`-like program, that shows how much time
is left. It is similar to my [`timer`](https://github.com/caarlos0/timer) TUI, if you’re interested in
something more complete.

Without further ado, let’s create the app.

First, navigate
[here](https://github.com/charmbracelet/bubbletea-app-template/generate) to
create a new repository based on our
[bubbletea-app-template](https://github.com/charmbracelet/bubbletea-app-template)
repository.

Then clone it. In my case, I called it `teatest-example`:

```
gh repo clone caarlos0/teatest-example
cd teatest-example
go mod tidy
$EDITOR .

```

This example will just sleep until the user presses `q` to exit.

With a few modifications we can get what we want:

- Add a `duration time.Duration` field to the `model`. We’ll use this to keep
track of how long we should sleep.
- Add a `start time.Time` field to the `model` to mark when we started the
countdown.
- The `initialModel` needs to take the `duration` as an argument. Set it into the
`model`, as well as setting `start` to `time.Now()`.
- Add a `timeLeft` method to the model, which calculates how long we still
need to sleep.
- In the `Update` method, we need to check if that `timeLeft > 0`, and quit
otherwise.
- In the `View` method, we need to display how much time is left.
- Finally, in `main` we parse `os.Args[1]` to a `time.Duration` and pass it down
to `initialModel`.

And that’s pretty much it. Here’s the
[link to the full diff](https://github.com/caarlos0/teatest-example/commit/ff0da36c7c68fef74cf173f5407e70b4347a178b).

## Imports

Before anything else, we need to import the `teatest` package:

```
go get github.com/charmbracelet/x/exp/teatest@latest

```

## The full output test

Next let’s create a `main_test.go` and start with a simple test that asserts
the entire final output of the app.

Here’s what it looks like:

```
// main_test.go
func TestFullOutput(t *testing.T) {
	m := initialModel(time.Second)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(300, 100))
	out, err := io.ReadAll(tm.FinalOutput(t))
	if err != nil {
		t.Error(err)
	}
	teatest.RequireEqualOutput(t, out)
}

```

1. We created a `model` that will sleep for 1 second.
2. We passed it to `teatest.NewTestModel`, also ensuring a fixed terminal size
3. We ask for the `FinalOutput`, and read it all.
4. `Final` means it will wait for the `tea.Program` to finish before
returning, so be wary that this will block until that condition is met.
5. We check if the output we got is equal the output in the _golden file_.

If you just run `go test ./...`, you’ll see that it errors. That’s because we
don’t have a golden file yet. To fix that, run:

```
go test -v ./... -update

```

The `-update` flag comes from the `teatest` package. It will update the
golden file (or create it if it doesn’t exist).

You can also `cat` the golden file to see what it looks like:

```
> cat testdata/TestFullOutput.golden

    ⣻  sleeping 0s... press q to quit

```

In subsequent tests, you’ll want to run `go test` without the `-update`, unless
you changed the output portion of your program.

Here’s the [link to the full diff](https://github.com/caarlos0/teatest-example/commit/c42a7180d7a9e39486315bcd9992182f201136a6).

## The final model test

Bubble Tea returns the final model after it finishes running, so we can also
assert against that final model:

```
// main_test.go
func TestFinalModel(t *testing.T) {
	tm := teatest.NewTestModel(t, initialModel(time.Second), teatest.WithInitialTermSize(300, 100))
	fm := tm.FinalModel(t)
	m, ok := fm.(model)
	if !ok {
		t.Fatalf("final model have the wrong type: %T", fm)
	}
	if m.duration != time.Second {
		t.Errorf("m.duration != 1s: %s", m.duration)
	}
	if m.start.After(time.Now().Add(-1 * time.Second)) {
		t.Errorf("m.start should be more than 1 second ago: %s", m.start)
	}
}

```

The setup is basically the same as the previous test, but instead of the asking
for the `FinalOutput`, we ask for the `FinalModel`.

We then need to cast it to the concrete type and then, finally, we assert for
the `m.duration` and `m.start`.

Here’s the [link to the full diff](https://github.com/caarlos0/teatest-example/commit/d474afa44676a3485623ce6b6966ca78fbdf4829).

## Intermediate output and sending messages

Another useful test case is to ensure things happen during the test.
We also need to interact with the program while its running.

Let’s write a quick test exploring these options:

```
// main_test.go
func TestOuput(t *testing.T) {
	tm := teatest.NewTestModel(t, initialModel(10*time.Second), teatest.WithInitialTermSize(300, 100))

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("sleeping 8s"))
	}, teatest.WithCheckInterval(time.Millisecond*100), teatest.WithDuration(time.Second*3))

	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("q"),
	})

	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

```

We setup our `teatest` in the same fashion as the previous test, then we assert
that the app, at some point, is showing `sleeping 8s`, meaning 2 seconds have
elapsed. We give that condition 3 seconds of time to be met, or else we fail.

Finally, we send a `tea.KeyMsg` with the character `q` on it, which should cause
the app to quit.

To ensure it quits in time, we `WaitFinished` with a timeout of 1 second.
This way we can be sure we finished because we send a `q` key press, not because
the program runs its 10 seconds out.

Here’s the [link to the full diff](https://github.com/caarlos0/teatest-example/commit/f63faaa3338af783ac40deb22e7196d7898d052a).

## The CI is failing. What now?

Once you push your commits GitHub Actions will test them and likely fail.

The reason for this is because your local golden file was generated with
whatever color profile the terminal `go test` was run in reported while GitHub
Actions is probably reporting something different.

Luckily, we can force everything to use the same color profile:

```
// main_test.go
func init() {
	lipgloss.SetColorProfile(termenv.Ascii)
}

```

In this app we don’t need to worry too much about colors, so its fine to use the
`Ascii` profile, which disables colors.

Another thing that might cause tests to fail is line endings. The golden files
look like text, but their line endings shouldn’t be messed with—and git might
just do that.

To remedy the situation, I recommend adding this to your `.gitattributes` file:

```gitattributes
*.golden -text

```

This will keep Git from handling them as text files.

Here’s the [link to the full diff](https://github.com/caarlos0/teatest-example/commit/93d97374a7c9c40711a342c3ab3cc51f3a5a58dd).

## Final words

This is an experimental, work in progress library, hence the
`github.com/charmbracelet/x/exp/teatest`
package name.

We encourage you to try it out in your projects and report back what you find.

And, if you’re interested, here’s the [link to the repository](https://github.com/caarlos0/teatest-example) for this
post.

EOF

Read this post in your terminal with [Glow](https://github.com/charmbracelet/glow):

`glow -p https://charm.land/blog/teatest.md` Copied!

By [Carlos Becker](https://github.com/caarlos0)

8 May 2023

Carlos writes and operates software for a living. He makes the command line glamourous at Charm and helps people release software faster with [GoReleaser](https://goreleaser.com/).

## Tail the logs.

Keep up with Charm and get access to our alphas, betas, experimental projects and stuff like that. We’re always building stuff and love getting feedback from talented people like you.

Submit

## Let’s chat!

Have a question about a command line thing you’re building? Got an idea for a new feature? Just wanna hang out? You’re always welcome in the [Charm Discord](https://charm.land/chat/).

- [172.0k](https://github.com/charmbracelet)

Copyright © 2025 Charmbracelet, Inc.

- [vt100@charm.land](mailto:vt100@charm.land)

Charm operates fully distributed from the United States, Brazil, Canada, Kosovo and Sweden.

haters > /dev/null™
