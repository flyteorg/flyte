import { Admin } from 'flyteidl';
import {
  Identifier,
  LiteralMap,
  Notification,
  ParameterMap,
  Schedule,
  VariableMap,
} from 'models/Common/types';

export interface LaunchPlanMetadata extends Admin.ILaunchPlanMetadata {
  notifications: Notification[];
  schedule: Schedule;
}

export interface LaunchPlanSpec extends Admin.ILaunchPlanSpec {
  defaultInputs: ParameterMap;
  entityMetadata: LaunchPlanMetadata;
  fixedInputs: LiteralMap;
  role: string;
  workflowId: Identifier;
}

export interface LaunchPlanClosure extends Admin.ILaunchPlanClosure {
  state: Admin.LaunchPlanState;
  expectedInputs: ParameterMap;
  expectedOutputs: VariableMap;
}

export interface LaunchPlan extends Admin.ILaunchPlan {
  closure?: LaunchPlanClosure;
  id: Identifier;
  spec: LaunchPlanSpec;
}

export type LaunchPlanId = Identifier;

/* It's an ENUM exports, and as such need to be exported as both type and const value */
/* eslint-disable @typescript-eslint/no-redeclare */
export type LaunchPlanState = Admin.LaunchPlanState;
export const LaunchPlanState = Admin.LaunchPlanState;
/* eslint-enable @typescript-eslint/no-redeclare */
