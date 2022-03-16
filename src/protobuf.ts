// Needed in order for protobufjs to properly convert int64s
import * as Long from 'long';
import * as $protobuf from 'protobufjs/minimal';

$protobuf.util.Long = Long;
$protobuf.configure();
