package com.witelokk.models

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class StartFollowingRequest(
    @SerialName("artist_id") val artistId: String,
)
