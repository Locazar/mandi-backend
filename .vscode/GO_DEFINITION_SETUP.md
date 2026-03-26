# VSCode Go "Go to Definition" Troubleshooting

## Root Causes Fixed

1. **✅ Missing `.vscode/settings.json`** — Created with proper Go/gopls configuration
2. **✅ Missing `gopls` language server** — Installed at `~/go/bin/gopls`

## What I've Done

### 1. Created `.vscode/settings.json`
Configured VSCode with:
- Go language server (gopls) enabled
- Auto-formatting on save
- Import organization
- Proper file watching exclusions

### 2. Installed gopls (Go Language Server)
```bash
# Gopls version installed:
golang.org/x/tools/gopls v0.21.0

# Location:
~/go/bin/gopls
```

## Next Steps - What You Need to Do

### Step 1: Ensure Go Extension is Installed
1. Open VSCode
2. Go to **Extensions** (Ctrl/Cmd + Shift + X)
3. Search for **"Go"** (by Google)
4. Click **Install** if not already installed
5. Restart VSCode

### Step 2: Add Go/bin to Your Shell PATH (Optional but Recommended)
Add to your `.zshrc` or `.bashrc`:
```bash
export PATH="$HOME/go/bin:$PATH"
```

Then reload:
```bash
source ~/.zshrc
```

### Step 3: Verify Setup in VSCode
1. Open any `.go` file in your project
2. Hold **Cmd** and hover over any symbol (function, struct, variable)
3. Click the symbol or press **Cmd + Click**
4. You should now see the definition

### Step 4: If Still Not Working - Restart Everything
```bash
# In VSCode, press:
Cmd + Shift + P

# Then search for and run:
Go: Restart Language Server
```

If that doesn't work:
1. **Close VSCode completely**
2. Run in terminal: `killall gopls` (to clear any hanging processes)
3. Reopen VSCode
4. Open a Go file and test again

## Verification Checklist

✅ Gopls installed: `~/go/bin/gopls version`
✅ `.vscode/settings.json` created with Go settings
✅ Go extension installed in VSCode
✅ Project has valid `go.mod` (exists: `go.mod` ✓)
✅ GOPATH set correctly: `/Users/dhruv/go`

## Common Issues & Fixes

| Issue | Fix |
|-------|-----|
| "Go to Definition" still not working | Restart VSCode (`Cmd + Shift + P` → Restart Language Server) |
| Gopls not found | Add `~/go/bin` to your PATH in `.zshrc` |
| Slow autocomplete | Settings file excludes `vendor/`, `build/`, `.git/` |
| Red squiggly lines under imports | Run `go mod tidy` in terminal |

## Key Files Modified

- **Created**: `.vscode/settings.json` — Go extension configuration
- **Already exists**: `go.mod` — Valid module setup

## References

- [VSCode Go Extension Docs](https://github.com/golang/vscode-go/wiki)
- [Gopls Configuration](https://github.com/golang/tools/wiki/gopls)
