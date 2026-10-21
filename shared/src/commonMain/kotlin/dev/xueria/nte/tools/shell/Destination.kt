package dev.xueria.nte.tools.shell

val AllDestinations: List<Destination> = listOf(
    Destination.MusicPlayer,
    Destination.Setting
)

val MainDestinations: List<Destination> = listOf(
    Destination.MusicPlayer
)

sealed class Destination(
    val id: String,
    val label: String,
    val title: String,
    val description: String
) {

    object MusicPlayer : Destination(
        id = "music_player",
        label = "MusicPlayer",
        title = "Music Player",
        description = "Auto play music in game."
    )

    object Setting : Destination(
        id = "setting",
        label = "Setting",
        title = "Setting",
        description = "Application setting."
    )
}