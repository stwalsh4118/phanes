# WSL Setup Guide for Vagrant Testing

This guide helps you set up Vagrant and VirtualBox to work from WSL.

## Quick Setup

### Option 1: Add to WSL PATH with .exe aliases (Recommended for WSL)

WSL sometimes requires `.exe` extensions. Add these lines to your `~/.zshrc`:

```bash
# Vagrant (adjust path if installed elsewhere)
export PATH="$PATH:/mnt/c/HashiCorp/Vagrant/bin"
# or if installed to Program Files:
# export PATH="$PATH:/mnt/c/Program Files/HashiCorp/Vagrant/bin"

# VirtualBox
export PATH="$PATH:/mnt/c/Program Files/Oracle/VirtualBox"

# WSL sometimes needs explicit .exe - create aliases
alias vagrant='vagrant.exe'
alias VBoxManage='VBoxManage.exe'
```

Then reload your shell:
```bash
source ~/.zshrc
```

### Option 2: Add to Windows PATH (Works everywhere)

1. Press `Win+R`, type `sysdm.cpl`, press Enter
2. Click **Advanced** tab
3. Click **Environment Variables**
4. Under **System variables**, find and select **Path**, then click **Edit**
5. Click **New** and add:
   - `C:\Program Files\Oracle\VirtualBox`
   - `C:\HashiCorp\Vagrant\bin` (or wherever Vagrant is installed)
6. Click **OK** on all dialogs
7. **Restart your WSL terminal** (close and reopen)

## Finding Installation Paths

### Vagrant
Common locations:
- `C:\HashiCorp\Vagrant\bin\vagrant.exe`
- `C:\Program Files\HashiCorp\Vagrant\bin\vagrant.exe`

Find it:
```bash
# From WSL
ls /mnt/c/HashiCorp/Vagrant/bin/vagrant.exe
ls /mnt/c/Program\ Files/HashiCorp/Vagrant/bin/vagrant.exe

# From Windows CMD
where vagrant
```

### VirtualBox
Common location:
- `C:\Program Files\Oracle\VirtualBox\VBoxManage.exe`

Find it:
```bash
# From WSL
ls "/mnt/c/Program Files/Oracle/VirtualBox/VBoxManage.exe"

# From Windows CMD
where VBoxManage
```

## Verification

After setup, verify from WSL:

```bash
# Check Vagrant
vagrant --version
# or
vagrant.exe --version

# Check VirtualBox
VBoxManage --version
# or
VBoxManage.exe --version
```

## Troubleshooting

### Vagrant requires .exe extension

**Symptom**: `vagrant --version` fails, but `vagrant.exe --version` works

**Solution**: Add alias to `~/.zshrc`:
```bash
alias vagrant='vagrant.exe'
```

Or always use `vagrant.exe` explicitly.

### VBoxManage not found

**Symptom**: `VBoxManage` doesn't work, even from Windows

**First, verify VirtualBox is installed correctly:**

1. **Check from Windows File Explorer:**
   - Navigate to `C:\Program Files\Oracle\VirtualBox`
   - Look for `VBoxManage.exe` file

2. **Test from Windows Command Prompt (cmd.exe):**
   ```cmd
   "C:\Program Files\Oracle\VirtualBox\VBoxManage.exe" --version
   ```
   
   If this fails, VirtualBox may not be fully installed. Try:
   - Reinstalling VirtualBox from https://www.virtualbox.org/wiki/Downloads
   - Make sure you install the full version, not just the GUI

3. **If VBoxManage.exe exists but doesn't work:**
   - VirtualBox may need to be repaired
   - Try running VirtualBox installer again and choose "Repair"

4. **For WSL access:**
   ```bash
   # Add to ~/.zshrc
   export PATH="$PATH:/mnt/c/Program Files/Oracle/VirtualBox"
   alias VBoxManage='VBoxManage.exe'
   
   # Test
   VBoxManage.exe --version
   ```

### "command not found" errors

1. **Check if files exist:**
   ```bash
   ls "/mnt/c/Program Files/Oracle/VirtualBox/VBoxManage.exe"
   ls /mnt/c/HashiCorp/Vagrant/bin/vagrant.exe
   ```

2. **If files exist but not in PATH:**
   - Use Option 1 above (add to WSL PATH with aliases)
   - Or use Option 2 (add to Windows PATH and restart WSL)

3. **If files don't exist:**
   - Reinstall Vagrant/VirtualBox
   - Check installation location during setup

### Script detects tools but they don't work

The script auto-detects Windows executables in common locations. If it finds them but they don't work:

1. Make sure you're running from WSL (not pure Linux)
2. Try running the Windows executables directly:
   ```bash
   "/mnt/c/Program Files/Oracle/VirtualBox/VBoxManage.exe" --version
   /mnt/c/HashiCorp/Vagrant/bin/vagrant.exe --version
   ```

3. If direct execution works, the script should work too

### Permission errors

If you see permission errors when running Windows executables from WSL:

1. Make sure VirtualBox and Vagrant are installed correctly
2. Try running from Windows Command Prompt first to verify installation
3. Some Windows executables may need to be run with `cmd.exe /c` wrapper

## Testing the Setup

Once configured, test the setup:

```bash
cd /home/sean/workspace/phanes
./scripts/test-vm.sh setup
```

If you see errors about missing tools, check the paths above and add them to your PATH.

