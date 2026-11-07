# This hook replaces the hook that linuxdeploy-plugin-qt writes. That hook asks for the
# platform theme "gtk2", which Qt6 does not have any more. Without a platform theme the
# application uses neither the colors nor the fonts of the desktop. The AppRun of
# linuxdeploy sources this file by its name, therefore the name must stay as it is.

case "${XDG_CURRENT_DESKTOP}" in
    *GNOME*|*gnome*|*XFCE*|*Cinnamon*|*MATE*|*Unity*)
        export QT_QPA_PLATFORMTHEME=gtk3
        ;;
esac
