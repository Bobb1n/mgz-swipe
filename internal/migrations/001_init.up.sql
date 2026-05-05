CREATE TABLE IF NOT EXISTS swipes (
    id         BIGSERIAL PRIMARY KEY,
    swiper_id  UUID        NOT NULL,
    swipee_id  UUID        NOT NULL,
    direction  VARCHAR(10) NOT NULL CHECK (direction IN ('like', 'dislike')),
    created_at TIMESTAMP   NOT NULL DEFAULT NOW(),
    UNIQUE (swiper_id, swipee_id)
);

CREATE INDEX IF NOT EXISTS idx_swipes_swiper_id  ON swipes(swiper_id);
CREATE INDEX IF NOT EXISTS idx_swipes_swipee_id  ON swipes(swipee_id);
CREATE INDEX IF NOT EXISTS idx_swipes_direction  ON swipes(swiper_id, direction);

CREATE TABLE IF NOT EXISTS matches (
    id         BIGSERIAL PRIMARY KEY,
    user1_id   UUID      NOT NULL,
    user2_id   UUID      NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (user1_id, user2_id)
);

CREATE INDEX IF NOT EXISTS idx_matches_user1 ON matches(user1_id);
CREATE INDEX IF NOT EXISTS idx_matches_user2 ON matches(user2_id);
