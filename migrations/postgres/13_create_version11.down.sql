-- Copyright 2020 Cray Inc. All Rights Reserved.
--
-- Except as permitted by contract or express written permission of Cray Inc.,
-- no part of this work or its content may be modified, used, reproduced or
-- disclosed in any form. Modifications made without express permission of
-- Cray Inc. may damage the system the software is installed within, may
-- disqualify the user from receiving support from Cray Inc. under support or
-- maintenance contracts, or require additional support services outside the
-- scope of those contracts to repair the software or system.

-- Completely remove schema 11 if we roll back this far.

BEGIN;

DROP TABLE IF EXISTS comp_eth_interfaces;

-- Decrease the schema version
INSERT INTO system VALUES(0, 10, '{}'::JSON)
    ON CONFLICT(id) DO UPDATE SET schema_version=10;

COMMIT;