package dev.xueria.nte.tools.platform

val JavaVersion: String
    get() = System.getProperty("java.version")

val RunningPath: String
    get() = System.getProperty("user.dir")