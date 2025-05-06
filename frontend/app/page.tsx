
import Navbar from "@/components/navbar"
import Footer from "@/components/footer"
import InfoSection from "@/components/InfoSection"
import MapSection from "@/components/MapSection"

export default function HomePage() {
  return (
    <div className="flex flex-col min-h-screen">
      <Navbar />
      <main className="flex-1 container mx-auto p-4">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="md:col-span-2">
            <MapSection />
          </div>
          <div className="md:col-span-1">
            <InfoSection />
          </div>
        </div>
      </main>
      <Footer />
    </div>
  )
}

