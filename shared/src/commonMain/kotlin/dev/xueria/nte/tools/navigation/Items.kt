package dev.xueria.nte.tools.navigation

val MusicPlayer: NavigationItem = NavigationItem(
    id = "music_player",
    label = "MusicPlayer",
    title = "Music Player",
    description = "why",
    icon = NavigationIcon.MusicPlayer
)

val DummyFeature: NavigationItem = NavigationItem(
    id = "dummy",
    label = "Dummy",
    title = "Dummy Title",
    description = "Dummy Description",
    icon = NavigationIcon.MusicPlayer
)

val NavigationItems = listOf(
    MusicPlayer,
    DummyFeature
)


