// Mock for ../wailsjs/go/models

class AutomationStepPayload {
  action: string = '';
  target: string = '';
  value: string = '';
  label: string = '';
  static createFrom(source: Record<string, unknown> = {}) { return new AutomationStepPayload(source); }
  constructor(source: Record<string, unknown> = {}) {
    this.action = typeof source['action'] === 'string' ? source['action'] : '';
    this.target = typeof source['target'] === 'string' ? source['target'] : '';
    this.value = typeof source['value'] === 'string' ? source['value'] : '';
    this.label = typeof source['label'] === 'string' ? source['label'] : '';
  }
}

class AutomationPayload {
  id: string = '';
  platformId: string = '';
  name: string = '';
  description: string = '';
  steps: AutomationStepPayload[] = [];
  createdAt: string = '';
  lastRun: string = '';
  runCount: number = 0;
  static createFrom(source: Record<string, unknown> = {}) { return new AutomationPayload(source); }
  constructor(source: Record<string, unknown> = {}) {
    this.id = typeof source['id'] === 'string' ? source['id'] : '';
    this.platformId = typeof source['platformId'] === 'string' ? source['platformId'] : '';
    this.name = typeof source['name'] === 'string' ? source['name'] : '';
    this.description = typeof source['description'] === 'string' ? source['description'] : '';
    this.steps = Array.isArray(source['steps'])
      ? (source['steps'] as Record<string, unknown>[]).map(s => AutomationStepPayload.createFrom(s))
      : [];
    this.createdAt = typeof source['createdAt'] === 'string' ? source['createdAt'] : '';
    this.lastRun = typeof source['lastRun'] === 'string' ? source['lastRun'] : '';
    this.runCount = typeof source['runCount'] === 'number' ? source['runCount'] : 0;
  }
}

export const main = {
  AutomationPayload,
  AutomationStepPayload,
};
