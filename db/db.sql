ALTER SYSTEM SET shared_buffers = '500 MB';
ALTER SYSTEM SET effective_cache_size = '1 GB';
ALTER SYSTEM SET work_mem = '32 MB';
ALTER SYSTEM SET max_wal_senders = 0;
ALTER SYSTEM SET wal_level = minimal;
ALTER SYSTEM SET fsync = OFF;
ALTER SYSTEM SET full_page_writes = OFF;
ALTER SYSTEM SET synchronous_commit = OFF;
ALTER SYSTEM SET archive_mode = OFF;

CREATE TYPE ENUM_VOICE AS ENUM ('1', '-1');

SELECT pg_reload_conf();

CREATE
    EXTENSION CITEXT;

CREATE
    UNLOGGED TABLE profile
(
    id       SERIAL,
    nickname CITEXT COLLATE "C" NOT NULL UNIQUE,
    fullname TEXT               NOT NULL,
    about    TEXT               NOT NULL DEFAULT '',
    email    CITEXT             NOT NULL UNIQUE,

    PRIMARY KEY (id)
);

CREATE INDEX ON profile (nickname);

CREATE
    UNLOGGED TABLE forum
(
    slug             CITEXT,
    title            TEXT   NOT NULL,
    profile_nickname CITEXT NOT NULL,
    posts            INT    NOT NULL DEFAULT 0,
    threads          INT    NOT NULL DEFAULT 0,

    PRIMARY KEY (slug),
    FOREIGN KEY (profile_nickname) REFERENCES profile (nickname) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX ON forum (slug);

CREATE
    UNLOGGED TABLE forum_user
(
    forum_slug       CITEXT             NOT NULL,
    profile_nickname CITEXT COLLATE "C" NOT NULL,
    profile_fullname TEXT               NOT NULL,
    profile_about    TEXT               NOT NULL,
    profile_email    CITEXT             NOT NULL,

    PRIMARY KEY (forum_slug, profile_nickname),
    FOREIGN KEY (forum_slug) REFERENCES forum (slug) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (profile_nickname) REFERENCES profile (nickname) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (profile_email) REFERENCES profile (email) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX ON forum_user (forum_slug);

CREATE
    UNLOGGED TABLE thread
(
    id               SERIAL,
    title            TEXT   NOT NULL,
    profile_nickname CITEXT NOT NULL,
    forum_slug       CITEXT NOT NULL,
    message          TEXT   NOT NULL,
    votes            INT    NOT NULL DEFAULT 0,
    slug             CITEXT UNIQUE,
    created          TIMESTAMPTZ,

    PRIMARY KEY (id),
    FOREIGN KEY (profile_nickname) REFERENCES profile (nickname) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (forum_slug) REFERENCES forum (slug) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX ON thread (slug);
CREATE INDEX ON thread (forum_slug);

CREATE
    UNLOGGED TABLE post
(
    id               BIGSERIAL,
    post_parent_id   BIGINT,
    profile_nickname CITEXT    NOT NULL,
    message          TEXT      NOT NULL,
    is_edited        BOOLEAN   NOT NULL DEFAULT false,
    forum_slug       CITEXT    NOT NULL,
    thread_id        INT       NOT NULL,
    created          TIMESTAMP NOT NULL,
    post_root_id     BIGINT    NOT NULL,
    path_            BIGINT[]  NOT NULL,

    PRIMARY KEY (id),
    FOREIGN KEY (profile_nickname) REFERENCES profile (nickname) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (post_root_id) REFERENCES post (id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (post_parent_id) REFERENCES post (id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (thread_id) REFERENCES thread (id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (forum_slug) REFERENCES forum (slug) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX ON post (thread_id);
CREATE INDEX ON post (thread_id, path_, created, id);
CREATE INDEX ON post (post_root_id);
CREATE INDEX ON post (post_root_id, path_, created, id);
CREATE INDEX ON post (thread_id, created, id);

CREATE
    UNLOGGED TABLE vote
(
    profile_id INT        NOT NULL,
    thread_id  INT        NOT NULL,
    voice      ENUM_VOICE NOT NULL,

    PRIMARY KEY (profile_id, thread_id),
    FOREIGN KEY (profile_id) REFERENCES profile (id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (thread_id) REFERENCES thread (id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX ON vote (profile_id);
CREATE INDEX ON vote (thread_id);



CREATE OR REPLACE FUNCTION USF_TRIGGER_thread_after_INSERT()
    RETURNS TRIGGER
AS
$$
BEGIN
    UPDATE forum
    SET threads = threads + 1
    WHERE forum.slug = NEW.forum_slug;
    INSERT INTO forum_user (forum_slug, profile_nickname, profile_about, profile_email, profile_fullname)
    SELECT NEW.forum_slug, NEW.profile_nickname, profile.about, profile.email, profile.fullname
    FROM profile
    WHERE profile.nickname = NEW.profile_nickname
    ON CONFLICT (forum_slug, profile_nickname) DO NOTHING;
    RETURN NEW;
END;
$$ LANGUAGE plpgSQL;

CREATE TRIGGER thread_after_INSERT
    AFTER INSERT
    ON thread
    FOR EACH ROW
EXECUTE PROCEDURE USF_TRIGGER_thread_after_INSERT();


CREATE OR REPLACE FUNCTION USF_TRIGGER_thread_before_INSERT()
    RETURNS TRIGGER
AS
$$
BEGIN
    IF
        NEW.slug = '' THEN
        NEW.slug := NULL;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgSQL;


CREATE TRIGGER thread_before_INSERT
    BEFORE INSERT
    ON thread
    FOR EACH ROW
EXECUTE PROCEDURE USF_TRIGGER_thread_before_INSERT();



CREATE OR REPLACE FUNCTION USF_TRIGGER_profile_after_UPDATE()
    RETURNS TRIGGER
AS
$$
BEGIN
    IF
                OLD.about != NEW.about OR OLD.fullname != NEW.fullname THEN
        UPDATE forum_user
        SET profile_about    = NEW.about,
            profile_fullname = NEW.fullname
        WHERE forum_user.profile_nickname = NEW.nickname;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgSQL;

CREATE TRIGGER profile_after_UPDATE
    AFTER INSERT
    ON profile
    FOR EACH ROW
EXECUTE PROCEDURE USF_TRIGGER_profile_after_UPDATE();




CREATE OR REPLACE FUNCTION USF_TRIGGER_post_before_UPDATE()
    RETURNS TRIGGER
AS
$$
BEGIN
    NEW.is_edited
        := TRUE;
    RETURN NEW;
END;
$$ LANGUAGE plpgSQL;


CREATE TRIGGER post_before_UPDATE
    BEFORE UPDATE
    ON post
    FOR EACH ROW
EXECUTE PROCEDURE USF_TRIGGER_post_before_UPDATE();



CREATE OR REPLACE FUNCTION USF_TRIGGER_post_after_INSERT()
    RETURNS TRIGGER
AS
$$
BEGIN
    UPDATE forum
    SET posts = posts + 1
    WHERE forum.slug = NEW.forum_slug;
    INSERT INTO forum_user (forum_slug, profile_nickname, profile_about, profile_email, profile_fullname)
    SELECT NEW.forum_slug, NEW.profile_nickname, profile.about, profile.email, profile.fullname
    FROM profile
    WHERE profile.nickname = NEW.profile_nickname
    ON CONFLICT (forum_slug, profile_nickname) DO NOTHING;
    RETURN NEW;
END;
$$ LANGUAGE plpgSQL;

CREATE TRIGGER post_after_INSERT
    AFTER INSERT
    ON post
    FOR EACH ROW
EXECUTE PROCEDURE USF_TRIGGER_post_after_INSERT();


CREATE OR REPLACE FUNCTION USF_TRIGGER_post_before_INSERT()
    RETURNS TRIGGER
AS
$$
BEGIN
    IF
        NEW.post_parent_id != 0 THEN
        NEW.path_ := (SELECT post.path_
                      FROM post
                      WHERE post.thread_id = NEW.thread_id
                        AND post.id = NEW.post_parent_id) || ARRAY [NEW.id];
        IF
            cardinality(NEW.path_) = 1 THEN
            RAISE 'Logic err: parent post in another thread';
        END IF;

        NEW.post_root_id
            := NEW.path_[1];
    ELSE
        NEW.post_parent_id := NULL;
        NEW.post_root_id
            := NEW.id;
        NEW.path_
            := ARRAY [NEW.id];
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgSQL;

CREATE TRIGGER post_before_INSERT
    BEFORE INSERT
    ON post
    FOR EACH ROW
EXECUTE PROCEDURE USF_TRIGGER_post_before_INSERT();


CREATE OR REPLACE FUNCTION USF_TRIGGER_vote_after_UPDATE()
    RETURNS TRIGGER
AS
$$
BEGIN
    IF
            OLD.voice != NEW.voice THEN
        IF NEW.voice = '1' THEN
            UPDATE thread
            SET votes = votes + 2
            WHERE thread.id = NEW.thread_id;
        ELSE
            UPDATE thread
            SET votes = votes - 2
            WHERE thread.id = NEW.thread_id;
        END IF;
    END IF;
    RETURN OLD;
END;
$$ LANGUAGE plpgSQL;

CREATE TRIGGER vote_after_UPDATE
    AFTER UPDATE
    ON vote
    FOR EACH ROW
EXECUTE PROCEDURE USF_TRIGGER_vote_after_UPDATE();

CREATE OR REPLACE FUNCTION USF_TRIGGER_vote_after_INSERT()
    RETURNS TRIGGER
AS
$$
BEGIN
    IF
        NEW.voice = '1' THEN
        UPDATE thread
        SET votes = votes + 1
        WHERE thread.id = NEW.thread_id;
    ELSE
        UPDATE thread
        SET votes = votes - 1
        WHERE thread.id = NEW.thread_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgSQL;

CREATE TRIGGER vote_after_INSERT
    AFTER INSERT
    ON vote
    FOR EACH ROW
EXECUTE PROCEDURE USF_TRIGGER_vote_after_INSERT();

