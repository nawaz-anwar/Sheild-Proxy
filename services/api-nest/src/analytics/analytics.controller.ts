import { Controller, Get, Query } from '@nestjs/common';
import { AnalyticsService } from './analytics.service';

@Controller('/analytics')
export class AnalyticsController {
  constructor(private readonly analyticsService: AnalyticsService) {}

  @Get('/overview')
  overview(@Query('domainId') domainId?: string) {
    return this.analyticsService.overview(domainId);
  }

  @Get('/time-series')
  timeSeries(@Query('domainId') domainId?: string, @Query('hours') hours?: string) {
    return this.analyticsService.timeSeries(domainId, Number(hours ?? 24));
  }

  @Get('/top-ips')
  topIPs(@Query('domainId') domainId?: string, @Query('limit') limit?: string) {
    return this.analyticsService.topIPs(domainId, Number(limit ?? 10));
  }
}
