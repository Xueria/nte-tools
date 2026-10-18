package dev.xueria.nte.tools.platform

actual val JavaVersion: String
    get() = System.getProperty("java.version")