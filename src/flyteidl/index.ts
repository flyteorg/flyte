import { flyteidl, google } from '@flyteorg/flyteidl/gen/pb-js/flyteidl';

/** Message classes for flyte entities */
import admin = flyteidl.admin;
import core = flyteidl.core;
import service = flyteidl.service;

/** Message classes for built-in Protobuf types */
import protobuf = google.protobuf;

export { admin as Admin, core as Core, service as Service, protobuf as Protobuf };
