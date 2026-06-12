import { z } from 'zod';

export const responseProtectionSettingsSchema = z.object({
  action: z.enum(['mask', 'reject', 'failover']),
  replacement: z.string().nullable().optional(),
  scopes: z.array(z.string()).default([]),
});

export const responseProtectionRuleSchema = z.object({
  id: z.string(),
  createdAt: z.coerce.date(),
  updatedAt: z.coerce.date(),
  name: z.string(),
  description: z.string(),
  pattern: z.string(),
  status: z.enum(['enabled', 'disabled', 'archived']),
  settings: responseProtectionSettingsSchema,
});

export const responseProtectionRuleEdgeSchema = z.object({
  node: responseProtectionRuleSchema,
  cursor: z.string(),
});

export const responseProtectionRulePageInfoSchema = z.object({
  hasNextPage: z.boolean(),
  hasPreviousPage: z.boolean(),
  startCursor: z.string().nullable(),
  endCursor: z.string().nullable(),
});

export const responseProtectionRuleConnectionSchema = z.object({
  edges: z.array(responseProtectionRuleEdgeSchema),
  pageInfo: responseProtectionRulePageInfoSchema,
  totalCount: z.number(),
});

export const responseProtectionRulePreviewResultSchema = z.object({
  result: z.string(),
  hasMatch: z.boolean(),
});

export type ResponseProtectionSettings = z.infer<typeof responseProtectionSettingsSchema>;
export type ResponseProtectionRule = z.infer<typeof responseProtectionRuleSchema>;
export type ResponseProtectionRuleConnection = z.infer<typeof responseProtectionRuleConnectionSchema>;
export type ResponseProtectionRulePreviewResult = z.infer<typeof responseProtectionRulePreviewResultSchema>;

export interface CreateResponseProtectionRuleInput {
  name: string;
  description?: string;
  pattern: string;
  settings: {
    action: 'mask' | 'reject' | 'failover';
    replacement?: string;
    scopes?: string[];
  };
}

export interface UpdateResponseProtectionRuleInput {
  name?: string;
  description?: string;
  pattern?: string;
  status?: 'enabled' | 'disabled' | 'archived';
  settings?: {
    action: 'mask' | 'reject' | 'failover';
    replacement?: string;
    scopes?: string[];
  };
}

export interface PreviewResponseProtectionRuleInput {
  pattern: string;
  testText: string;
  settings: {
    action: 'mask' | 'reject' | 'failover';
    replacement?: string;
    scopes?: string[];
  };
}
