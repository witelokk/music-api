package com.witelokk.music.tables

import org.jetbrains.exposed.sql.Table

object ReleaseArtists: Table("release_artists") {
    val releaseId = uuid("release_id").references(Releases.id)
    val artistId = uuid("artist_id").references(Artists.id)

    override val primaryKey = PrimaryKey(releaseId, artistId)
}