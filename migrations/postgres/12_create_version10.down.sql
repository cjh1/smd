-- Copyright 2020 Cray Inc. All Rights Reserved.
--
-- Except as permitted by contract or express written permission of Cray Inc.,
-- no part of this work or its content may be modified, used, reproduced or
-- disclosed in any form. Modifications made without express permission of
-- Cray Inc. may damage the system the software is installed within, may
-- disqualify the user from receiving support from Cray Inc. under support or
-- maintenance contracts, or require additional support services outside the
-- scope of those contracts to repair the software or system.

-- Completely remove schema 10 if we roll back this far.

BEGIN;

DROP VIEW IF EXISTS hwinv_by_loc_with_fru;

ALTER TABLE hwinv_by_loc
ALTER COLUMN "fru_id" TYPE varchar(63);

ALTER TABLE hwinv_by_fru
ALTER COLUMN "fru_id" TYPE VARCHAR(128);

ALTER TABLE hwinv_hist
ALTER COLUMN "fru_id" TYPE VARCHAR(128);

CREATE OR REPLACE VIEW hwinv_by_loc_with_fru AS
SELECT
    hwinv_by_loc.id             AS  "id",
    hwinv_by_loc.type           AS  "type",
    hwinv_by_loc.ordinal        AS  "ordinal",
    hwinv_by_loc.status         AS  "status",
    hwinv_by_loc.location_info  AS  "location_info", -- JSON blob
    hwinv_by_loc.fru_id         AS  "fru_id",
    hwinv_by_fru.type           AS  "fru_type",
    hwinv_by_fru.subtype        AS  "fru_subtype",
    hwinv_by_fru.fru_info       AS  "fru_info"       -- JSON blob
FROM hwinv_by_loc
LEFT JOIN hwinv_by_fru ON hwinv_by_loc.fru_id = hwinv_by_fru.fru_id;

-- Decrease the schema version
INSERT INTO system VALUES(0, 9, '{}'::JSON)
    ON CONFLICT(id) DO UPDATE SET schema_version=9;

COMMIT;