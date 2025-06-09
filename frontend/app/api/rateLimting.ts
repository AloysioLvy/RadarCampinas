import {Ratelimit} from "@upstash/ratelimit";
import {Redis} from "@upstash/redis";
import { NextResponse } from "next/server";
 
 
const redis = new Redis({
  url: process.env.UPSTASH_REDIS_REST_URL,
  token:  process.env.UPSTASH_REDIS_REST_TOKEN,
})
 
// Create a new ratelimiter, that allows 80 requests per IP per day
export const rateLimit = new Ratelimit({
  redis: redis,
  limiter: Ratelimit.slidingWindow(20, "1 d"),
  analytics:true

});
 
export async function checkRateLimit(req: Request) {
    const ip = req.headers.get("x-forwarded-for") || "anonymous";
    const result = await rateLimit.limit(ip);
    const headers = new Headers();
    headers.set("X-RateLimit-Limit", result.limit.toString());
    headers.set("X-RateLimit-Remaining", result.remaining.toString());
    headers.set("X-RateLimit-Reset", result.reset.toString());

  if (!result.success) {
    return NextResponse.json(
      {
        message: "Too many requests",
        rateLimitState: result,
      },
      {
        status: 429,
        headers,
      }
    );
  }

  return { success: true, headers };
}