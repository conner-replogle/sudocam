export const enum RecordingType {
  RECORDING_TYPE_UNSPECIFIED = "RECORDING_TYPE_UNSPECIFIED",
  RECORDING_TYPE_OFF = "RECORDING_TYPE_OFF",
  RECORDING_TYPE_CONTINUOUS = "RECORDING_TYPE_CONTINUOUS",
  RECORDING_TYPE_CONTINUOUS_SCHEDULED = "RECORDING_TYPE_CONTINUOUS_SCHEDULED",
  RECORDING_TYPE_MOTION = "RECORDING_TYPE_MOTION",
}

export const encodeRecordingType: { [key: string]: number } = {
  RECORDING_TYPE_UNSPECIFIED: 0,
  RECORDING_TYPE_OFF: 1,
  RECORDING_TYPE_CONTINUOUS: 2,
  RECORDING_TYPE_CONTINUOUS_SCHEDULED: 3,
  RECORDING_TYPE_MOTION: 4,
};

export const decodeRecordingType: { [key: number]: RecordingType } = {
  0: RecordingType.RECORDING_TYPE_UNSPECIFIED,
  1: RecordingType.RECORDING_TYPE_OFF,
  2: RecordingType.RECORDING_TYPE_CONTINUOUS,
  3: RecordingType.RECORDING_TYPE_CONTINUOUS_SCHEDULED,
  4: RecordingType.RECORDING_TYPE_MOTION,
};

export interface Message {
  to?: string;
  from?: string;
  webrtc?: Webrtc;
  initalization?: Initalization;
  response?: Response;
  hls_request?: HLSRequest;
  hls_response?: HLSResponse;
  record_request?: RecordRequest;
  record_response?: RecordResponse;
  user_config?: UserConfig;
  trigger_refresh?: TriggerRefresh;
}

export function encodeMessage(message: Message): Uint8Array {
  let bb = popByteBuffer();
  _encodeMessage(message, bb);
  return toUint8Array(bb);
}

function _encodeMessage(message: Message, bb: ByteBuffer): void {
  // optional string to = 1;
  let $to = message.to;
  if ($to !== undefined) {
    writeVarint32(bb, 10);
    writeString(bb, $to);
  }

  // optional string from = 2;
  let $from = message.from;
  if ($from !== undefined) {
    writeVarint32(bb, 18);
    writeString(bb, $from);
  }

  // optional Webrtc webrtc = 3;
  let $webrtc = message.webrtc;
  if ($webrtc !== undefined) {
    writeVarint32(bb, 26);
    let nested = popByteBuffer();
    _encodeWebrtc($webrtc, nested);
    writeVarint32(bb, nested.limit);
    writeByteBuffer(bb, nested);
    pushByteBuffer(nested);
  }

  // optional Initalization initalization = 4;
  let $initalization = message.initalization;
  if ($initalization !== undefined) {
    writeVarint32(bb, 34);
    let nested = popByteBuffer();
    _encodeInitalization($initalization, nested);
    writeVarint32(bb, nested.limit);
    writeByteBuffer(bb, nested);
    pushByteBuffer(nested);
  }

  // optional Response response = 5;
  let $response = message.response;
  if ($response !== undefined) {
    writeVarint32(bb, 42);
    let nested = popByteBuffer();
    _encodeResponse($response, nested);
    writeVarint32(bb, nested.limit);
    writeByteBuffer(bb, nested);
    pushByteBuffer(nested);
  }

  // optional HLSRequest hls_request = 6;
  let $hls_request = message.hls_request;
  if ($hls_request !== undefined) {
    writeVarint32(bb, 50);
    let nested = popByteBuffer();
    _encodeHLSRequest($hls_request, nested);
    writeVarint32(bb, nested.limit);
    writeByteBuffer(bb, nested);
    pushByteBuffer(nested);
  }

  // optional HLSResponse hls_response = 7;
  let $hls_response = message.hls_response;
  if ($hls_response !== undefined) {
    writeVarint32(bb, 58);
    let nested = popByteBuffer();
    _encodeHLSResponse($hls_response, nested);
    writeVarint32(bb, nested.limit);
    writeByteBuffer(bb, nested);
    pushByteBuffer(nested);
  }

  // optional RecordRequest record_request = 8;
  let $record_request = message.record_request;
  if ($record_request !== undefined) {
    writeVarint32(bb, 66);
    let nested = popByteBuffer();
    _encodeRecordRequest($record_request, nested);
    writeVarint32(bb, nested.limit);
    writeByteBuffer(bb, nested);
    pushByteBuffer(nested);
  }

  // optional RecordResponse record_response = 9;
  let $record_response = message.record_response;
  if ($record_response !== undefined) {
    writeVarint32(bb, 74);
    let nested = popByteBuffer();
    _encodeRecordResponse($record_response, nested);
    writeVarint32(bb, nested.limit);
    writeByteBuffer(bb, nested);
    pushByteBuffer(nested);
  }

  // optional UserConfig user_config = 10;
  let $user_config = message.user_config;
  if ($user_config !== undefined) {
    writeVarint32(bb, 82);
    let nested = popByteBuffer();
    _encodeUserConfig($user_config, nested);
    writeVarint32(bb, nested.limit);
    writeByteBuffer(bb, nested);
    pushByteBuffer(nested);
  }

  // optional TriggerRefresh trigger_refresh = 11;
  let $trigger_refresh = message.trigger_refresh;
  if ($trigger_refresh !== undefined) {
    writeVarint32(bb, 90);
    let nested = popByteBuffer();
    _encodeTriggerRefresh($trigger_refresh, nested);
    writeVarint32(bb, nested.limit);
    writeByteBuffer(bb, nested);
    pushByteBuffer(nested);
  }
}

export function decodeMessage(binary: Uint8Array): Message {
  return _decodeMessage(wrapByteBuffer(binary));
}

function _decodeMessage(bb: ByteBuffer): Message {
  let message: Message = {} as any;

  end_of_message: while (!isAtEnd(bb)) {
    let tag = readVarint32(bb);

    switch (tag >>> 3) {
      case 0:
        break end_of_message;

      // optional string to = 1;
      case 1: {
        message.to = readString(bb, readVarint32(bb));
        break;
      }

      // optional string from = 2;
      case 2: {
        message.from = readString(bb, readVarint32(bb));
        break;
      }

      // optional Webrtc webrtc = 3;
      case 3: {
        let limit = pushTemporaryLength(bb);
        message.webrtc = _decodeWebrtc(bb);
        bb.limit = limit;
        break;
      }

      // optional Initalization initalization = 4;
      case 4: {
        let limit = pushTemporaryLength(bb);
        message.initalization = _decodeInitalization(bb);
        bb.limit = limit;
        break;
      }

      // optional Response response = 5;
      case 5: {
        let limit = pushTemporaryLength(bb);
        message.response = _decodeResponse(bb);
        bb.limit = limit;
        break;
      }

      // optional HLSRequest hls_request = 6;
      case 6: {
        let limit = pushTemporaryLength(bb);
        message.hls_request = _decodeHLSRequest(bb);
        bb.limit = limit;
        break;
      }

      // optional HLSResponse hls_response = 7;
      case 7: {
        let limit = pushTemporaryLength(bb);
        message.hls_response = _decodeHLSResponse(bb);
        bb.limit = limit;
        break;
      }

      // optional RecordRequest record_request = 8;
      case 8: {
        let limit = pushTemporaryLength(bb);
        message.record_request = _decodeRecordRequest(bb);
        bb.limit = limit;
        break;
      }

      // optional RecordResponse record_response = 9;
      case 9: {
        let limit = pushTemporaryLength(bb);
        message.record_response = _decodeRecordResponse(bb);
        bb.limit = limit;
        break;
      }

      // optional UserConfig user_config = 10;
      case 10: {
        let limit = pushTemporaryLength(bb);
        message.user_config = _decodeUserConfig(bb);
        bb.limit = limit;
        break;
      }

      // optional TriggerRefresh trigger_refresh = 11;
      case 11: {
        let limit = pushTemporaryLength(bb);
        message.trigger_refresh = _decodeTriggerRefresh(bb);
        bb.limit = limit;
        break;
      }

      default:
        skipUnknownField(bb, tag & 7);
    }
  }

  return message;
}

export interface HLSRequest {
  file_name?: string;
}

export function encodeHLSRequest(message: HLSRequest): Uint8Array {
  let bb = popByteBuffer();
  _encodeHLSRequest(message, bb);
  return toUint8Array(bb);
}

function _encodeHLSRequest(message: HLSRequest, bb: ByteBuffer): void {
  // optional string file_name = 2;
  let $file_name = message.file_name;
  if ($file_name !== undefined) {
    writeVarint32(bb, 18);
    writeString(bb, $file_name);
  }
}

export function decodeHLSRequest(binary: Uint8Array): HLSRequest {
  return _decodeHLSRequest(wrapByteBuffer(binary));
}

function _decodeHLSRequest(bb: ByteBuffer): HLSRequest {
  let message: HLSRequest = {} as any;

  end_of_message: while (!isAtEnd(bb)) {
    let tag = readVarint32(bb);

    switch (tag >>> 3) {
      case 0:
        break end_of_message;

      // optional string file_name = 2;
      case 2: {
        message.file_name = readString(bb, readVarint32(bb));
        break;
      }

      default:
        skipUnknownField(bb, tag & 7);
    }
  }

  return message;
}

export interface HLSResponse {
  file_name?: string;
  data?: Uint8Array;
}

export function encodeHLSResponse(message: HLSResponse): Uint8Array {
  let bb = popByteBuffer();
  _encodeHLSResponse(message, bb);
  return toUint8Array(bb);
}

function _encodeHLSResponse(message: HLSResponse, bb: ByteBuffer): void {
  // optional string file_name = 2;
  let $file_name = message.file_name;
  if ($file_name !== undefined) {
    writeVarint32(bb, 18);
    writeString(bb, $file_name);
  }

  // optional bytes data = 1;
  let $data = message.data;
  if ($data !== undefined) {
    writeVarint32(bb, 10);
    writeVarint32(bb, $data.length), writeBytes(bb, $data);
  }
}

export function decodeHLSResponse(binary: Uint8Array): HLSResponse {
  return _decodeHLSResponse(wrapByteBuffer(binary));
}

function _decodeHLSResponse(bb: ByteBuffer): HLSResponse {
  let message: HLSResponse = {} as any;

  end_of_message: while (!isAtEnd(bb)) {
    let tag = readVarint32(bb);

    switch (tag >>> 3) {
      case 0:
        break end_of_message;

      // optional string file_name = 2;
      case 2: {
        message.file_name = readString(bb, readVarint32(bb));
        break;
      }

      // optional bytes data = 1;
      case 1: {
        message.data = readBytes(bb, readVarint32(bb));
        break;
      }

      default:
        skipUnknownField(bb, tag & 7);
    }
  }

  return message;
}

export interface RecordRequest {
  id?: Long;
  start_time?: Long;
  end_time?: Long;
}

export function encodeRecordRequest(message: RecordRequest): Uint8Array {
  let bb = popByteBuffer();
  _encodeRecordRequest(message, bb);
  return toUint8Array(bb);
}

function _encodeRecordRequest(message: RecordRequest, bb: ByteBuffer): void {
  // optional int64 id = 1;
  let $id = message.id;
  if ($id !== undefined) {
    writeVarint32(bb, 8);
    writeVarint64(bb, $id);
  }

  // optional int64 start_time = 2;
  let $start_time = message.start_time;
  if ($start_time !== undefined) {
    writeVarint32(bb, 16);
    writeVarint64(bb, $start_time);
  }

  // optional int64 end_time = 3;
  let $end_time = message.end_time;
  if ($end_time !== undefined) {
    writeVarint32(bb, 24);
    writeVarint64(bb, $end_time);
  }
}

export function decodeRecordRequest(binary: Uint8Array): RecordRequest {
  return _decodeRecordRequest(wrapByteBuffer(binary));
}

function _decodeRecordRequest(bb: ByteBuffer): RecordRequest {
  let message: RecordRequest = {} as any;

  end_of_message: while (!isAtEnd(bb)) {
    let tag = readVarint32(bb);

    switch (tag >>> 3) {
      case 0:
        break end_of_message;

      // optional int64 id = 1;
      case 1: {
        message.id = readVarint64(bb, /* unsigned */ false);
        break;
      }

      // optional int64 start_time = 2;
      case 2: {
        message.start_time = readVarint64(bb, /* unsigned */ false);
        break;
      }

      // optional int64 end_time = 3;
      case 3: {
        message.end_time = readVarint64(bb, /* unsigned */ false);
        break;
      }

      default:
        skipUnknownField(bb, tag & 7);
    }
  }

  return message;
}

export interface VideoRange {
  start_time?: Long;
  end_time?: Long;
  file_name?: string;
}

export function encodeVideoRange(message: VideoRange): Uint8Array {
  let bb = popByteBuffer();
  _encodeVideoRange(message, bb);
  return toUint8Array(bb);
}

function _encodeVideoRange(message: VideoRange, bb: ByteBuffer): void {
  // optional int64 start_time = 1;
  let $start_time = message.start_time;
  if ($start_time !== undefined) {
    writeVarint32(bb, 8);
    writeVarint64(bb, $start_time);
  }

  // optional int64 end_time = 2;
  let $end_time = message.end_time;
  if ($end_time !== undefined) {
    writeVarint32(bb, 16);
    writeVarint64(bb, $end_time);
  }

  // optional string file_name = 3;
  let $file_name = message.file_name;
  if ($file_name !== undefined) {
    writeVarint32(bb, 26);
    writeString(bb, $file_name);
  }
}

export function decodeVideoRange(binary: Uint8Array): VideoRange {
  return _decodeVideoRange(wrapByteBuffer(binary));
}

function _decodeVideoRange(bb: ByteBuffer): VideoRange {
  let message: VideoRange = {} as any;

  end_of_message: while (!isAtEnd(bb)) {
    let tag = readVarint32(bb);

    switch (tag >>> 3) {
      case 0:
        break end_of_message;

      // optional int64 start_time = 1;
      case 1: {
        message.start_time = readVarint64(bb, /* unsigned */ false);
        break;
      }

      // optional int64 end_time = 2;
      case 2: {
        message.end_time = readVarint64(bb, /* unsigned */ false);
        break;
      }

      // optional string file_name = 3;
      case 3: {
        message.file_name = readString(bb, readVarint32(bb));
        break;
      }

      default:
        skipUnknownField(bb, tag & 7);
    }
  }

  return message;
}

export interface RecordResponse {
  id?: Long;
  records?: VideoRange[];
}

export function encodeRecordResponse(message: RecordResponse): Uint8Array {
  let bb = popByteBuffer();
  _encodeRecordResponse(message, bb);
  return toUint8Array(bb);
}

function _encodeRecordResponse(message: RecordResponse, bb: ByteBuffer): void {
  // optional int64 id = 1;
  let $id = message.id;
  if ($id !== undefined) {
    writeVarint32(bb, 8);
    writeVarint64(bb, $id);
  }

  // repeated VideoRange records = 2;
  let array$records = message.records;
  if (array$records !== undefined) {
    for (let value of array$records) {
      writeVarint32(bb, 18);
      let nested = popByteBuffer();
      _encodeVideoRange(value, nested);
      writeVarint32(bb, nested.limit);
      writeByteBuffer(bb, nested);
      pushByteBuffer(nested);
    }
  }
}

export function decodeRecordResponse(binary: Uint8Array): RecordResponse {
  return _decodeRecordResponse(wrapByteBuffer(binary));
}

function _decodeRecordResponse(bb: ByteBuffer): RecordResponse {
  let message: RecordResponse = {} as any;

  end_of_message: while (!isAtEnd(bb)) {
    let tag = readVarint32(bb);

    switch (tag >>> 3) {
      case 0:
        break end_of_message;

      // optional int64 id = 1;
      case 1: {
        message.id = readVarint64(bb, /* unsigned */ false);
        break;
      }

      // repeated VideoRange records = 2;
      case 2: {
        let limit = pushTemporaryLength(bb);
        let values = message.records || (message.records = []);
        values.push(_decodeVideoRange(bb));
        bb.limit = limit;
        break;
      }

      default:
        skipUnknownField(bb, tag & 7);
    }
  }

  return message;
}

export interface TriggerRefresh {
}

export function encodeTriggerRefresh(message: TriggerRefresh): Uint8Array {
  let bb = popByteBuffer();
  _encodeTriggerRefresh(message, bb);
  return toUint8Array(bb);
}

function _encodeTriggerRefresh(message: TriggerRefresh, bb: ByteBuffer): void {
}

export function decodeTriggerRefresh(binary: Uint8Array): TriggerRefresh {
  return _decodeTriggerRefresh(wrapByteBuffer(binary));
}

function _decodeTriggerRefresh(bb: ByteBuffer): TriggerRefresh {
  let message: TriggerRefresh = {} as any;

  end_of_message: while (!isAtEnd(bb)) {
    let tag = readVarint32(bb);

    switch (tag >>> 3) {
      case 0:
        break end_of_message;

      default:
        skipUnknownField(bb, tag & 7);
    }
  }

  return message;
}

export interface Webrtc {
  stream_id?: string;
  data?: string;
}

export function encodeWebrtc(message: Webrtc): Uint8Array {
  let bb = popByteBuffer();
  _encodeWebrtc(message, bb);
  return toUint8Array(bb);
}

function _encodeWebrtc(message: Webrtc, bb: ByteBuffer): void {
  // optional string stream_id = 1;
  let $stream_id = message.stream_id;
  if ($stream_id !== undefined) {
    writeVarint32(bb, 10);
    writeString(bb, $stream_id);
  }

  // optional string data = 2;
  let $data = message.data;
  if ($data !== undefined) {
    writeVarint32(bb, 18);
    writeString(bb, $data);
  }
}

export function decodeWebrtc(binary: Uint8Array): Webrtc {
  return _decodeWebrtc(wrapByteBuffer(binary));
}

function _decodeWebrtc(bb: ByteBuffer): Webrtc {
  let message: Webrtc = {} as any;

  end_of_message: while (!isAtEnd(bb)) {
    let tag = readVarint32(bb);

    switch (tag >>> 3) {
      case 0:
        break end_of_message;

      // optional string stream_id = 1;
      case 1: {
        message.stream_id = readString(bb, readVarint32(bb));
        break;
      }

      // optional string data = 2;
      case 2: {
        message.data = readString(bb, readVarint32(bb));
        break;
      }

      default:
        skipUnknownField(bb, tag & 7);
    }
  }

  return message;
}

export interface Initalization {
  id?: string;
  is_user?: boolean;
  token?: string;
}

export function encodeInitalization(message: Initalization): Uint8Array {
  let bb = popByteBuffer();
  _encodeInitalization(message, bb);
  return toUint8Array(bb);
}

function _encodeInitalization(message: Initalization, bb: ByteBuffer): void {
  // optional string id = 1;
  let $id = message.id;
  if ($id !== undefined) {
    writeVarint32(bb, 10);
    writeString(bb, $id);
  }

  // optional bool is_user = 2;
  let $is_user = message.is_user;
  if ($is_user !== undefined) {
    writeVarint32(bb, 16);
    writeByte(bb, $is_user ? 1 : 0);
  }

  // optional string token = 3;
  let $token = message.token;
  if ($token !== undefined) {
    writeVarint32(bb, 26);
    writeString(bb, $token);
  }
}

export function decodeInitalization(binary: Uint8Array): Initalization {
  return _decodeInitalization(wrapByteBuffer(binary));
}

function _decodeInitalization(bb: ByteBuffer): Initalization {
  let message: Initalization = {} as any;

  end_of_message: while (!isAtEnd(bb)) {
    let tag = readVarint32(bb);

    switch (tag >>> 3) {
      case 0:
        break end_of_message;

      // optional string id = 1;
      case 1: {
        message.id = readString(bb, readVarint32(bb));
        break;
      }

      // optional bool is_user = 2;
      case 2: {
        message.is_user = !!readByte(bb);
        break;
      }

      // optional string token = 3;
      case 3: {
        message.token = readString(bb, readVarint32(bb));
        break;
      }

      default:
        skipUnknownField(bb, tag & 7);
    }
  }

  return message;
}

export interface Response {
  message?: string;
  success?: boolean;
}

export function encodeResponse(message: Response): Uint8Array {
  let bb = popByteBuffer();
  _encodeResponse(message, bb);
  return toUint8Array(bb);
}

function _encodeResponse(message: Response, bb: ByteBuffer): void {
  // optional string message = 1;
  let $message = message.message;
  if ($message !== undefined) {
    writeVarint32(bb, 10);
    writeString(bb, $message);
  }

  // optional bool success = 2;
  let $success = message.success;
  if ($success !== undefined) {
    writeVarint32(bb, 16);
    writeByte(bb, $success ? 1 : 0);
  }
}

export function decodeResponse(binary: Uint8Array): Response {
  return _decodeResponse(wrapByteBuffer(binary));
}

function _decodeResponse(bb: ByteBuffer): Response {
  let message: Response = {} as any;

  end_of_message: while (!isAtEnd(bb)) {
    let tag = readVarint32(bb);

    switch (tag >>> 3) {
      case 0:
        break end_of_message;

      // optional string message = 1;
      case 1: {
        message.message = readString(bb, readVarint32(bb));
        break;
      }

      // optional bool success = 2;
      case 2: {
        message.success = !!readByte(bb);
        break;
      }

      default:
        skipUnknownField(bb, tag & 7);
    }
  }

  return message;
}

export interface Schedule {
  days_of_week?: number[];
  start_time?: string;
  end_time?: string;
}

export function encodeSchedule(message: Schedule): Uint8Array {
  let bb = popByteBuffer();
  _encodeSchedule(message, bb);
  return toUint8Array(bb);
}

function _encodeSchedule(message: Schedule, bb: ByteBuffer): void {
  // repeated int32 days_of_week = 1;
  let array$days_of_week = message.days_of_week;
  if (array$days_of_week !== undefined) {
    let packed = popByteBuffer();
    for (let value of array$days_of_week) {
      writeVarint64(packed, intToLong(value));
    }
    writeVarint32(bb, 10);
    writeVarint32(bb, packed.offset);
    writeByteBuffer(bb, packed);
    pushByteBuffer(packed);
  }

  // optional string start_time = 2;
  let $start_time = message.start_time;
  if ($start_time !== undefined) {
    writeVarint32(bb, 18);
    writeString(bb, $start_time);
  }

  // optional string end_time = 3;
  let $end_time = message.end_time;
  if ($end_time !== undefined) {
    writeVarint32(bb, 26);
    writeString(bb, $end_time);
  }
}

export function decodeSchedule(binary: Uint8Array): Schedule {
  return _decodeSchedule(wrapByteBuffer(binary));
}

function _decodeSchedule(bb: ByteBuffer): Schedule {
  let message: Schedule = {} as any;

  end_of_message: while (!isAtEnd(bb)) {
    let tag = readVarint32(bb);

    switch (tag >>> 3) {
      case 0:
        break end_of_message;

      // repeated int32 days_of_week = 1;
      case 1: {
        let values = message.days_of_week || (message.days_of_week = []);
        if ((tag & 7) === 2) {
          let outerLimit = pushTemporaryLength(bb);
          while (!isAtEnd(bb)) {
            values.push(readVarint32(bb));
          }
          bb.limit = outerLimit;
        } else {
          values.push(readVarint32(bb));
        }
        break;
      }

      // optional string start_time = 2;
      case 2: {
        message.start_time = readString(bb, readVarint32(bb));
        break;
      }

      // optional string end_time = 3;
      case 3: {
        message.end_time = readString(bb, readVarint32(bb));
        break;
      }

      default:
        skipUnknownField(bb, tag & 7);
    }
  }

  return message;
}

export interface MotionConfig {
  sensitivity?: number;
  pre_record_seconds?: number;
  post_record_seconds?: number;
}

export function encodeMotionConfig(message: MotionConfig): Uint8Array {
  let bb = popByteBuffer();
  _encodeMotionConfig(message, bb);
  return toUint8Array(bb);
}

function _encodeMotionConfig(message: MotionConfig, bb: ByteBuffer): void {
  // optional int32 sensitivity = 1;
  let $sensitivity = message.sensitivity;
  if ($sensitivity !== undefined) {
    writeVarint32(bb, 8);
    writeVarint64(bb, intToLong($sensitivity));
  }

  // optional int32 pre_record_seconds = 2;
  let $pre_record_seconds = message.pre_record_seconds;
  if ($pre_record_seconds !== undefined) {
    writeVarint32(bb, 16);
    writeVarint64(bb, intToLong($pre_record_seconds));
  }

  // optional int32 post_record_seconds = 3;
  let $post_record_seconds = message.post_record_seconds;
  if ($post_record_seconds !== undefined) {
    writeVarint32(bb, 24);
    writeVarint64(bb, intToLong($post_record_seconds));
  }
}

export function decodeMotionConfig(binary: Uint8Array): MotionConfig {
  return _decodeMotionConfig(wrapByteBuffer(binary));
}

function _decodeMotionConfig(bb: ByteBuffer): MotionConfig {
  let message: MotionConfig = {} as any;

  end_of_message: while (!isAtEnd(bb)) {
    let tag = readVarint32(bb);

    switch (tag >>> 3) {
      case 0:
        break end_of_message;

      // optional int32 sensitivity = 1;
      case 1: {
        message.sensitivity = readVarint32(bb);
        break;
      }

      // optional int32 pre_record_seconds = 2;
      case 2: {
        message.pre_record_seconds = readVarint32(bb);
        break;
      }

      // optional int32 post_record_seconds = 3;
      case 3: {
        message.post_record_seconds = readVarint32(bb);
        break;
      }

      default:
        skipUnknownField(bb, tag & 7);
    }
  }

  return message;
}

export interface UserConfig {
  recording_type?: RecordingType;
  schedules?: Schedule[];
  motion_config?: MotionConfig;
  motion_enabled?: boolean;
  name?: string;
}

export function encodeUserConfig(message: UserConfig): Uint8Array {
  let bb = popByteBuffer();
  _encodeUserConfig(message, bb);
  return toUint8Array(bb);
}

function _encodeUserConfig(message: UserConfig, bb: ByteBuffer): void {
  // optional RecordingType recording_type = 1;
  let $recording_type = message.recording_type;
  if ($recording_type !== undefined) {
    writeVarint32(bb, 8);
    writeVarint32(bb, encodeRecordingType[$recording_type]);
  }

  // repeated Schedule schedules = 2;
  let array$schedules = message.schedules;
  if (array$schedules !== undefined) {
    for (let value of array$schedules) {
      writeVarint32(bb, 18);
      let nested = popByteBuffer();
      _encodeSchedule(value, nested);
      writeVarint32(bb, nested.limit);
      writeByteBuffer(bb, nested);
      pushByteBuffer(nested);
    }
  }

  // optional MotionConfig motion_config = 3;
  let $motion_config = message.motion_config;
  if ($motion_config !== undefined) {
    writeVarint32(bb, 26);
    let nested = popByteBuffer();
    _encodeMotionConfig($motion_config, nested);
    writeVarint32(bb, nested.limit);
    writeByteBuffer(bb, nested);
    pushByteBuffer(nested);
  }

  // optional bool motion_enabled = 4;
  let $motion_enabled = message.motion_enabled;
  if ($motion_enabled !== undefined) {
    writeVarint32(bb, 32);
    writeByte(bb, $motion_enabled ? 1 : 0);
  }

  // optional string name = 5;
  let $name = message.name;
  if ($name !== undefined) {
    writeVarint32(bb, 42);
    writeString(bb, $name);
  }
}

export function decodeUserConfig(binary: Uint8Array): UserConfig {
  return _decodeUserConfig(wrapByteBuffer(binary));
}

function _decodeUserConfig(bb: ByteBuffer): UserConfig {
  let message: UserConfig = {} as any;

  end_of_message: while (!isAtEnd(bb)) {
    let tag = readVarint32(bb);

    switch (tag >>> 3) {
      case 0:
        break end_of_message;

      // optional RecordingType recording_type = 1;
      case 1: {
        message.recording_type = decodeRecordingType[readVarint32(bb)];
        break;
      }

      // repeated Schedule schedules = 2;
      case 2: {
        let limit = pushTemporaryLength(bb);
        let values = message.schedules || (message.schedules = []);
        values.push(_decodeSchedule(bb));
        bb.limit = limit;
        break;
      }

      // optional MotionConfig motion_config = 3;
      case 3: {
        let limit = pushTemporaryLength(bb);
        message.motion_config = _decodeMotionConfig(bb);
        bb.limit = limit;
        break;
      }

      // optional bool motion_enabled = 4;
      case 4: {
        message.motion_enabled = !!readByte(bb);
        break;
      }

      // optional string name = 5;
      case 5: {
        message.name = readString(bb, readVarint32(bb));
        break;
      }

      default:
        skipUnknownField(bb, tag & 7);
    }
  }

  return message;
}

export interface Timestamp {
  seconds?: Long;
  nanos?: number;
}

export function encodeTimestamp(message: Timestamp): Uint8Array {
  let bb = popByteBuffer();
  _encodeTimestamp(message, bb);
  return toUint8Array(bb);
}

function _encodeTimestamp(message: Timestamp, bb: ByteBuffer): void {
  // optional int64 seconds = 1;
  let $seconds = message.seconds;
  if ($seconds !== undefined) {
    writeVarint32(bb, 8);
    writeVarint64(bb, $seconds);
  }

  // optional int32 nanos = 2;
  let $nanos = message.nanos;
  if ($nanos !== undefined) {
    writeVarint32(bb, 16);
    writeVarint64(bb, intToLong($nanos));
  }
}

export function decodeTimestamp(binary: Uint8Array): Timestamp {
  return _decodeTimestamp(wrapByteBuffer(binary));
}

function _decodeTimestamp(bb: ByteBuffer): Timestamp {
  let message: Timestamp = {} as any;

  end_of_message: while (!isAtEnd(bb)) {
    let tag = readVarint32(bb);

    switch (tag >>> 3) {
      case 0:
        break end_of_message;

      // optional int64 seconds = 1;
      case 1: {
        message.seconds = readVarint64(bb, /* unsigned */ false);
        break;
      }

      // optional int32 nanos = 2;
      case 2: {
        message.nanos = readVarint32(bb);
        break;
      }

      default:
        skipUnknownField(bb, tag & 7);
    }
  }

  return message;
}

export interface Long {
  low: number;
  high: number;
  unsigned: boolean;
}

interface ByteBuffer {
  bytes: Uint8Array;
  offset: number;
  limit: number;
}

function pushTemporaryLength(bb: ByteBuffer): number {
  let length = readVarint32(bb);
  let limit = bb.limit;
  bb.limit = bb.offset + length;
  return limit;
}

function skipUnknownField(bb: ByteBuffer, type: number): void {
  switch (type) {
    case 0: while (readByte(bb) & 0x80) { } break;
    case 2: skip(bb, readVarint32(bb)); break;
    case 5: skip(bb, 4); break;
    case 1: skip(bb, 8); break;
    default: throw new Error("Unimplemented type: " + type);
  }
}

function stringToLong(value: string): Long {
  return {
    low: value.charCodeAt(0) | (value.charCodeAt(1) << 16),
    high: value.charCodeAt(2) | (value.charCodeAt(3) << 16),
    unsigned: false,
  };
}

function longToString(value: Long): string {
  let low = value.low;
  let high = value.high;
  return String.fromCharCode(
    low & 0xFFFF,
    low >>> 16,
    high & 0xFFFF,
    high >>> 16);
}

// The code below was modified from https://github.com/protobufjs/bytebuffer.js
// which is under the Apache License 2.0.

let f32 = new Float32Array(1);
let f32_u8 = new Uint8Array(f32.buffer);

let f64 = new Float64Array(1);
let f64_u8 = new Uint8Array(f64.buffer);

function intToLong(value: number): Long {
  value |= 0;
  return {
    low: value,
    high: value >> 31,
    unsigned: value >= 0,
  };
}

let bbStack: ByteBuffer[] = [];

function popByteBuffer(): ByteBuffer {
  const bb = bbStack.pop();
  if (!bb) return { bytes: new Uint8Array(64), offset: 0, limit: 0 };
  bb.offset = bb.limit = 0;
  return bb;
}

function pushByteBuffer(bb: ByteBuffer): void {
  bbStack.push(bb);
}

function wrapByteBuffer(bytes: Uint8Array): ByteBuffer {
  return { bytes, offset: 0, limit: bytes.length };
}

function toUint8Array(bb: ByteBuffer): Uint8Array {
  let bytes = bb.bytes;
  let limit = bb.limit;
  return bytes.length === limit ? bytes : bytes.subarray(0, limit);
}

function skip(bb: ByteBuffer, offset: number): void {
  if (bb.offset + offset > bb.limit) {
    throw new Error('Skip past limit');
  }
  bb.offset += offset;
}

function isAtEnd(bb: ByteBuffer): boolean {
  return bb.offset >= bb.limit;
}

function grow(bb: ByteBuffer, count: number): number {
  let bytes = bb.bytes;
  let offset = bb.offset;
  let limit = bb.limit;
  let finalOffset = offset + count;
  if (finalOffset > bytes.length) {
    let newBytes = new Uint8Array(finalOffset * 2);
    newBytes.set(bytes);
    bb.bytes = newBytes;
  }
  bb.offset = finalOffset;
  if (finalOffset > limit) {
    bb.limit = finalOffset;
  }
  return offset;
}

function advance(bb: ByteBuffer, count: number): number {
  let offset = bb.offset;
  if (offset + count > bb.limit) {
    throw new Error('Read past limit');
  }
  bb.offset += count;
  return offset;
}

function readBytes(bb: ByteBuffer, count: number): Uint8Array {
  let offset = advance(bb, count);
  return bb.bytes.subarray(offset, offset + count);
}

function writeBytes(bb: ByteBuffer, buffer: Uint8Array): void {
  let offset = grow(bb, buffer.length);
  bb.bytes.set(buffer, offset);
}

function readString(bb: ByteBuffer, count: number): string {
  // Sadly a hand-coded UTF8 decoder is much faster than subarray+TextDecoder in V8
  let offset = advance(bb, count);
  let fromCharCode = String.fromCharCode;
  let bytes = bb.bytes;
  let invalid = '\uFFFD';
  let text = '';

  for (let i = 0; i < count; i++) {
    let c1 = bytes[i + offset], c2: number, c3: number, c4: number, c: number;

    // 1 byte
    if ((c1 & 0x80) === 0) {
      text += fromCharCode(c1);
    }

    // 2 bytes
    else if ((c1 & 0xE0) === 0xC0) {
      if (i + 1 >= count) text += invalid;
      else {
        c2 = bytes[i + offset + 1];
        if ((c2 & 0xC0) !== 0x80) text += invalid;
        else {
          c = ((c1 & 0x1F) << 6) | (c2 & 0x3F);
          if (c < 0x80) text += invalid;
          else {
            text += fromCharCode(c);
            i++;
          }
        }
      }
    }

    // 3 bytes
    else if ((c1 & 0xF0) == 0xE0) {
      if (i + 2 >= count) text += invalid;
      else {
        c2 = bytes[i + offset + 1];
        c3 = bytes[i + offset + 2];
        if (((c2 | (c3 << 8)) & 0xC0C0) !== 0x8080) text += invalid;
        else {
          c = ((c1 & 0x0F) << 12) | ((c2 & 0x3F) << 6) | (c3 & 0x3F);
          if (c < 0x0800 || (c >= 0xD800 && c <= 0xDFFF)) text += invalid;
          else {
            text += fromCharCode(c);
            i += 2;
          }
        }
      }
    }

    // 4 bytes
    else if ((c1 & 0xF8) == 0xF0) {
      if (i + 3 >= count) text += invalid;
      else {
        c2 = bytes[i + offset + 1];
        c3 = bytes[i + offset + 2];
        c4 = bytes[i + offset + 3];
        if (((c2 | (c3 << 8) | (c4 << 16)) & 0xC0C0C0) !== 0x808080) text += invalid;
        else {
          c = ((c1 & 0x07) << 0x12) | ((c2 & 0x3F) << 0x0C) | ((c3 & 0x3F) << 0x06) | (c4 & 0x3F);
          if (c < 0x10000 || c > 0x10FFFF) text += invalid;
          else {
            c -= 0x10000;
            text += fromCharCode((c >> 10) + 0xD800, (c & 0x3FF) + 0xDC00);
            i += 3;
          }
        }
      }
    }

    else text += invalid;
  }

  return text;
}

function writeString(bb: ByteBuffer, text: string): void {
  // Sadly a hand-coded UTF8 encoder is much faster than TextEncoder+set in V8
  let n = text.length;
  let byteCount = 0;

  // Write the byte count first
  for (let i = 0; i < n; i++) {
    let c = text.charCodeAt(i);
    if (c >= 0xD800 && c <= 0xDBFF && i + 1 < n) {
      c = (c << 10) + text.charCodeAt(++i) - 0x35FDC00;
    }
    byteCount += c < 0x80 ? 1 : c < 0x800 ? 2 : c < 0x10000 ? 3 : 4;
  }
  writeVarint32(bb, byteCount);

  let offset = grow(bb, byteCount);
  let bytes = bb.bytes;

  // Then write the bytes
  for (let i = 0; i < n; i++) {
    let c = text.charCodeAt(i);
    if (c >= 0xD800 && c <= 0xDBFF && i + 1 < n) {
      c = (c << 10) + text.charCodeAt(++i) - 0x35FDC00;
    }
    if (c < 0x80) {
      bytes[offset++] = c;
    } else {
      if (c < 0x800) {
        bytes[offset++] = ((c >> 6) & 0x1F) | 0xC0;
      } else {
        if (c < 0x10000) {
          bytes[offset++] = ((c >> 12) & 0x0F) | 0xE0;
        } else {
          bytes[offset++] = ((c >> 18) & 0x07) | 0xF0;
          bytes[offset++] = ((c >> 12) & 0x3F) | 0x80;
        }
        bytes[offset++] = ((c >> 6) & 0x3F) | 0x80;
      }
      bytes[offset++] = (c & 0x3F) | 0x80;
    }
  }
}

function writeByteBuffer(bb: ByteBuffer, buffer: ByteBuffer): void {
  let offset = grow(bb, buffer.limit);
  let from = bb.bytes;
  let to = buffer.bytes;

  // This for loop is much faster than subarray+set on V8
  for (let i = 0, n = buffer.limit; i < n; i++) {
    from[i + offset] = to[i];
  }
}

function readByte(bb: ByteBuffer): number {
  return bb.bytes[advance(bb, 1)];
}

function writeByte(bb: ByteBuffer, value: number): void {
  let offset = grow(bb, 1);
  bb.bytes[offset] = value;
}

function readFloat(bb: ByteBuffer): number {
  let offset = advance(bb, 4);
  let bytes = bb.bytes;

  // Manual copying is much faster than subarray+set in V8
  f32_u8[0] = bytes[offset++];
  f32_u8[1] = bytes[offset++];
  f32_u8[2] = bytes[offset++];
  f32_u8[3] = bytes[offset++];
  return f32[0];
}

function writeFloat(bb: ByteBuffer, value: number): void {
  let offset = grow(bb, 4);
  let bytes = bb.bytes;
  f32[0] = value;

  // Manual copying is much faster than subarray+set in V8
  bytes[offset++] = f32_u8[0];
  bytes[offset++] = f32_u8[1];
  bytes[offset++] = f32_u8[2];
  bytes[offset++] = f32_u8[3];
}

function readDouble(bb: ByteBuffer): number {
  let offset = advance(bb, 8);
  let bytes = bb.bytes;

  // Manual copying is much faster than subarray+set in V8
  f64_u8[0] = bytes[offset++];
  f64_u8[1] = bytes[offset++];
  f64_u8[2] = bytes[offset++];
  f64_u8[3] = bytes[offset++];
  f64_u8[4] = bytes[offset++];
  f64_u8[5] = bytes[offset++];
  f64_u8[6] = bytes[offset++];
  f64_u8[7] = bytes[offset++];
  return f64[0];
}

function writeDouble(bb: ByteBuffer, value: number): void {
  let offset = grow(bb, 8);
  let bytes = bb.bytes;
  f64[0] = value;

  // Manual copying is much faster than subarray+set in V8
  bytes[offset++] = f64_u8[0];
  bytes[offset++] = f64_u8[1];
  bytes[offset++] = f64_u8[2];
  bytes[offset++] = f64_u8[3];
  bytes[offset++] = f64_u8[4];
  bytes[offset++] = f64_u8[5];
  bytes[offset++] = f64_u8[6];
  bytes[offset++] = f64_u8[7];
}

function readInt32(bb: ByteBuffer): number {
  let offset = advance(bb, 4);
  let bytes = bb.bytes;
  return (
    bytes[offset] |
    (bytes[offset + 1] << 8) |
    (bytes[offset + 2] << 16) |
    (bytes[offset + 3] << 24)
  );
}

function writeInt32(bb: ByteBuffer, value: number): void {
  let offset = grow(bb, 4);
  let bytes = bb.bytes;
  bytes[offset] = value;
  bytes[offset + 1] = value >> 8;
  bytes[offset + 2] = value >> 16;
  bytes[offset + 3] = value >> 24;
}

function readInt64(bb: ByteBuffer, unsigned: boolean): Long {
  return {
    low: readInt32(bb),
    high: readInt32(bb),
    unsigned,
  };
}

function writeInt64(bb: ByteBuffer, value: Long): void {
  writeInt32(bb, value.low);
  writeInt32(bb, value.high);
}

function readVarint32(bb: ByteBuffer): number {
  let c = 0;
  let value = 0;
  let b: number;
  do {
    b = readByte(bb);
    if (c < 32) value |= (b & 0x7F) << c;
    c += 7;
  } while (b & 0x80);
  return value;
}

function writeVarint32(bb: ByteBuffer, value: number): void {
  value >>>= 0;
  while (value >= 0x80) {
    writeByte(bb, (value & 0x7f) | 0x80);
    value >>>= 7;
  }
  writeByte(bb, value);
}

function readVarint64(bb: ByteBuffer, unsigned: boolean): Long {
  let part0 = 0;
  let part1 = 0;
  let part2 = 0;
  let b: number;

  b = readByte(bb); part0 = (b & 0x7F); if (b & 0x80) {
    b = readByte(bb); part0 |= (b & 0x7F) << 7; if (b & 0x80) {
      b = readByte(bb); part0 |= (b & 0x7F) << 14; if (b & 0x80) {
        b = readByte(bb); part0 |= (b & 0x7F) << 21; if (b & 0x80) {

          b = readByte(bb); part1 = (b & 0x7F); if (b & 0x80) {
            b = readByte(bb); part1 |= (b & 0x7F) << 7; if (b & 0x80) {
              b = readByte(bb); part1 |= (b & 0x7F) << 14; if (b & 0x80) {
                b = readByte(bb); part1 |= (b & 0x7F) << 21; if (b & 0x80) {

                  b = readByte(bb); part2 = (b & 0x7F); if (b & 0x80) {
                    b = readByte(bb); part2 |= (b & 0x7F) << 7;
                  }
                }
              }
            }
          }
        }
      }
    }
  }

  return {
    low: part0 | (part1 << 28),
    high: (part1 >>> 4) | (part2 << 24),
    unsigned,
  };
}

function writeVarint64(bb: ByteBuffer, value: Long): void {
  let part0 = value.low >>> 0;
  let part1 = ((value.low >>> 28) | (value.high << 4)) >>> 0;
  let part2 = value.high >>> 24;

  // ref: src/google/protobuf/io/coded_stream.cc
  let size =
    part2 === 0 ?
      part1 === 0 ?
        part0 < 1 << 14 ?
          part0 < 1 << 7 ? 1 : 2 :
          part0 < 1 << 21 ? 3 : 4 :
        part1 < 1 << 14 ?
          part1 < 1 << 7 ? 5 : 6 :
          part1 < 1 << 21 ? 7 : 8 :
      part2 < 1 << 7 ? 9 : 10;

  let offset = grow(bb, size);
  let bytes = bb.bytes;

  switch (size) {
    case 10: bytes[offset + 9] = (part2 >>> 7) & 0x01;
    case 9: bytes[offset + 8] = size !== 9 ? part2 | 0x80 : part2 & 0x7F;
    case 8: bytes[offset + 7] = size !== 8 ? (part1 >>> 21) | 0x80 : (part1 >>> 21) & 0x7F;
    case 7: bytes[offset + 6] = size !== 7 ? (part1 >>> 14) | 0x80 : (part1 >>> 14) & 0x7F;
    case 6: bytes[offset + 5] = size !== 6 ? (part1 >>> 7) | 0x80 : (part1 >>> 7) & 0x7F;
    case 5: bytes[offset + 4] = size !== 5 ? part1 | 0x80 : part1 & 0x7F;
    case 4: bytes[offset + 3] = size !== 4 ? (part0 >>> 21) | 0x80 : (part0 >>> 21) & 0x7F;
    case 3: bytes[offset + 2] = size !== 3 ? (part0 >>> 14) | 0x80 : (part0 >>> 14) & 0x7F;
    case 2: bytes[offset + 1] = size !== 2 ? (part0 >>> 7) | 0x80 : (part0 >>> 7) & 0x7F;
    case 1: bytes[offset] = size !== 1 ? part0 | 0x80 : part0 & 0x7F;
  }
}

function readVarint32ZigZag(bb: ByteBuffer): number {
  let value = readVarint32(bb);

  // ref: src/google/protobuf/wire_format_lite.h
  return (value >>> 1) ^ -(value & 1);
}

function writeVarint32ZigZag(bb: ByteBuffer, value: number): void {
  // ref: src/google/protobuf/wire_format_lite.h
  writeVarint32(bb, (value << 1) ^ (value >> 31));
}

function readVarint64ZigZag(bb: ByteBuffer): Long {
  let value = readVarint64(bb, /* unsigned */ false);
  let low = value.low;
  let high = value.high;
  let flip = -(low & 1);

  // ref: src/google/protobuf/wire_format_lite.h
  return {
    low: ((low >>> 1) | (high << 31)) ^ flip,
    high: (high >>> 1) ^ flip,
    unsigned: false,
  };
}

function writeVarint64ZigZag(bb: ByteBuffer, value: Long): void {
  let low = value.low;
  let high = value.high;
  let flip = high >> 31;

  // ref: src/google/protobuf/wire_format_lite.h
  writeVarint64(bb, {
    low: (low << 1) ^ flip,
    high: ((high << 1) | (low >>> 31)) ^ flip,
    unsigned: false,
  });
}
