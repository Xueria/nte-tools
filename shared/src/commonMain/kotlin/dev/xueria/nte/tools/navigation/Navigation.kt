package dev.xueria.nte.tools.navigation

enum class NavigationIcon {
    MusicPlayer
}

data class NavigationItem(
    val id: String,
    val label: String,
    val title: String,
    val description: String,
    val icon: NavigationIcon
)