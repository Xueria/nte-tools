package dev.xueria.nte.tools.json

import kotlinx.serialization.json.Json

const val JsonFileExtension = "json"

val GlobalJson = Json {
    ignoreUnknownKeys = true
    prettyPrint = true
}