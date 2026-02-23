import HomeChat from './HomeChat'
import { config } from '../config'
import './Home.css'

export default function Home() {
  return (
    <section className="home">
      <h2 className="home-greeting">Hi, I'm <span className="home-accent">{config.SITE_FIRST_NAME}</span></h2>
      <p className="home-summary">
        Product leader with 15+ years building data and AI-powered platforms
        that turn behavioural signals into trusted, high-performing user
        experiences. I lead multi-PM organisations and partner deeply with
        Data Science and Engineering to move from research to
        production — shipping intelligent products that improve matching,
        discovery, decisioning, and outcomes at scale.
      </p>
      <HomeChat />
    </section>
  )
}
