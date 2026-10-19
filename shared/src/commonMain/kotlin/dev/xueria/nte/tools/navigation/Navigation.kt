package dev.xueria.nte.tools.navigation

import dev.xueria.nte.tools.language.TranslateKey

enum class NavigationIcon {
    MusicPlayer
}

data class NavigationItem(
    val id: String,
    val label: String,
    val title: String,
    val description: String,
    val icon: NavigationIcon
) {

    val labelTranslateKey = TranslateKey("navigation").dot(id).dot("label")
    val titleTranslateKey = TranslateKey("navigation").dot(id).dot("title")
    val descriptionTranslateKey = TranslateKey("navigation").dot(id).dot("description")

}