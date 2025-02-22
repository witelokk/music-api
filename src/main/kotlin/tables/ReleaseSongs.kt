package com.witelokk.tables

import org.jetbrains.exposed.sql.Table

object ReleaseSongs: Table("release_songs") {
    val releaseId = uuid("releaseId").references(Releases.id)
    val songId = uuid("songId").references(Songs.id)

    override val primaryKey = PrimaryKey(releaseId, songId)
}