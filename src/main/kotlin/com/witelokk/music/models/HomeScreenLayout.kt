package com.witelokk.music.models

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class HomeScreenLayout(
    val playlists: PlaylistsSummary,
    val followedArtists: ArtistsSummary,
    val sections: List<Section>,
) {
    companion object {
        @Serializable
        data class Section(
            val title: String,
            @SerialName("title_ru") val titleRu: String,
            val releases: Releases,
        )
        @Serializable
        data class Sections(
            val count: Int,
            val sections: List<Section>
        )
    }
}