#!/bin/bash
# Diagnostic script to check Vagrant and VirtualBox setup

echo "🔍 Checking Vagrant and VirtualBox setup..."
echo ""

# Check if running in WSL
if [ -d "/mnt/c" ] && [ ! -f "/.dockerenv" ]; then
    echo "✓ Running in WSL"
    IS_WSL=true
else
    echo "✓ Running in native Linux"
    IS_WSL=false
fi
echo ""

# Check Vagrant
echo "📦 Checking Vagrant..."
VAGRANT_FOUND=false

# Try without .exe
if command -v vagrant &>/dev/null; then
    echo "  ✓ Found: vagrant (command)"
    vagrant --version 2>&1 | head -1 | sed 's/^/    /'
    VAGRANT_FOUND=true
fi

# Try with .exe
if command -v vagrant.exe &>/dev/null; then
    echo "  ✓ Found: vagrant.exe (command)"
    vagrant.exe --version 2>&1 | head -1 | sed 's/^/    /'
    VAGRANT_FOUND=true
fi

# Check common Windows paths
if [ "$IS_WSL" = true ]; then
    for path in \
        "/mnt/c/HashiCorp/Vagrant/bin/vagrant.exe" \
        "/mnt/c/Program Files/HashiCorp/Vagrant/bin/vagrant.exe"
    do
        if [ -f "$path" ]; then
            echo "  ✓ Found: $path"
            "$path" --version 2>&1 | head -1 | sed 's/^/    /'
            VAGRANT_FOUND=true
        fi
    done
fi

if [ "$VAGRANT_FOUND" = false ]; then
    echo "  ✗ Vagrant not found"
    echo "    Install from: https://www.vagrantup.com/downloads"
fi
echo ""

# Check VirtualBox
echo "📦 Checking VirtualBox VBoxManage..."
VBOX_FOUND=false

# Try without .exe
if command -v VBoxManage &>/dev/null; then
    echo "  ✓ Found: VBoxManage (command)"
    VBoxManage --version 2>&1 | head -1 | sed 's/^/    /'
    VBOX_FOUND=true
fi

# Try with .exe
if command -v VBoxManage.exe &>/dev/null; then
    echo "  ✓ Found: VBoxManage.exe (command)"
    VBoxManage.exe --version 2>&1 | head -1 | sed 's/^/    /'
    VBOX_FOUND=true
fi

# Check common Windows paths
if [ "$IS_WSL" = true ]; then
    for path in \
        "/mnt/c/Program Files/Oracle/VirtualBox/VBoxManage.exe" \
        "/mnt/c/Program Files (x86)/Oracle/VirtualBox/VBoxManage.exe"
    do
        if [ -f "$path" ]; then
            echo "  ✓ Found: $path"
            if "$path" --version &>/dev/null; then
                "$path" --version 2>&1 | head -1 | sed 's/^/    /'
                VBOX_FOUND=true
            else
                echo "    ⚠ File exists but doesn't execute properly"
                echo "    This may indicate VirtualBox needs repair/reinstall"
            fi
        fi
    done
fi

if [ "$VBOX_FOUND" = false ]; then
    echo "  ✗ VBoxManage not found or not working"
    echo ""
    echo "  Troubleshooting steps:"
    echo "  1. Verify VirtualBox is installed:"
    echo "     - Check: C:\\Program Files\\Oracle\\VirtualBox\\VBoxManage.exe"
    echo ""
    echo "  2. Test from Windows Command Prompt:"
    echo "     \"C:\\Program Files\\Oracle\\VirtualBox\\VBoxManage.exe\" --version"
    echo ""
    echo "  3. If it doesn't work from Windows, VirtualBox may need repair:"
    echo "     - Reinstall from: https://www.virtualbox.org/wiki/Downloads"
    echo ""
    if [ "$IS_WSL" = true ]; then
        echo "  4. For WSL access, add to ~/.zshrc:"
        echo "     export PATH=\"\$PATH:/mnt/c/Program Files/Oracle/VirtualBox\""
        echo "     alias VBoxManage='VBoxManage.exe'"
    fi
fi
echo ""

# Summary
echo "📋 Summary:"
if [ "$VAGRANT_FOUND" = true ] && [ "$VBOX_FOUND" = true ]; then
    echo "  ✅ Both tools are accessible!"
    echo "  You should be able to run: ./scripts/test-vm.sh setup"
elif [ "$VAGRANT_FOUND" = true ]; then
    echo "  ⚠ Vagrant found, but VirtualBox VBoxManage is missing"
    echo "  Fix VirtualBox before proceeding"
elif [ "$VBOX_FOUND" = true ]; then
    echo "  ⚠ VirtualBox found, but Vagrant is missing"
    echo "  Install Vagrant before proceeding"
else
    echo "  ❌ Both tools are missing or not accessible"
    echo "  Install both Vagrant and VirtualBox before proceeding"
fi


